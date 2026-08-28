<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, RefreshCw } from '@lucide/vue'

const route = useRoute()
const router = useRouter()
const serverId = computed(() => route.params.serverId)
const containerName = computed(() => typeof route.query.container === 'string' ? route.query.container : '')

const terminalEl = ref(null)
const status = ref('connecting')
const statusMessage = ref('正在连接…')
const initError = ref('')

let term = null
let fitAddon = null
let ws = null

async function initTerminal() {
  try {
    const [{ Terminal }, { FitAddon }] = await Promise.all([
      import('@xterm/xterm'),
      import('@xterm/addon-fit'),
      import('@xterm/xterm/css/xterm.css')
    ])

    term = new Terminal({
      fontFamily: '"SFMono-Regular", Menlo, Monaco, Consolas, monospace',
      fontSize: 13,
      lineHeight: 1.45,
      letterSpacing: 0,
      cursorBlink: true,
      cursorStyle: 'bar',
      scrollback: 5000,
      screenReaderMode: true,
      theme: {
        background: '#101114',
        foreground: '#e5e5ea',
        cursor: '#66a8ff',
        cursorAccent: '#101114',
        selectionBackground: 'rgba(0, 113, 227, 0.42)',
        black: '#1d1d1f',
        red: '#ff6961',
        green: '#4dcc77',
        yellow: '#ffbd45',
        blue: '#66a8ff',
        magenta: '#bf8cff',
        cyan: '#5ac8d8',
        white: '#d1d1d6',
        brightBlack: '#6e6e73',
        brightRed: '#ff817a',
        brightGreen: '#65db8d',
        brightYellow: '#ffd06a',
        brightBlue: '#86bbff',
        brightMagenta: '#d0a9ff',
        brightCyan: '#7ad7e4',
        brightWhite: '#f5f5f7'
      }
    })

    fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(terminalEl.value)
    fitAddon.fit()

    term.onData(data => {
      if (ws?.readyState === WebSocket.OPEN) ws.send(data)
    })

    term.onResize(({ cols, rows }) => {
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'resize', cols, rows }))
      }
    })

    connectWebSocket()
  } catch (error) {
    initError.value = `终端无法启动：${error.message}`
    status.value = 'disconnected'
    statusMessage.value = '终端初始化失败'
  }
}

function connectWebSocket() {
  const token = encodeURIComponent(localStorage.getItem('token') || '')
  const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
  const encodedServerId = encodeURIComponent(serverId.value)
  const path = containerName.value
    ? `/api/terminal/container/ws/${encodedServerId}/${encodeURIComponent(containerName.value)}`
    : `/api/terminal/ws/${encodedServerId}`
  const socket = new WebSocket(`${protocol}://${location.host}${path}?token=${token}`)
  ws = socket

  socket.onopen = () => {
    if (ws !== socket) return
    status.value = 'connected'
    statusMessage.value = containerName.value ? `已连接容器 ${containerName.value}` : '服务器终端已连接'
    fitAddon?.fit()
  }

  socket.onmessage = event => {
    if (ws === socket) term?.write(event.data)
  }

  socket.onclose = () => {
    if (ws !== socket) return
    status.value = 'disconnected'
    statusMessage.value = '连接已断开'
    term?.write('\r\n\x1b[31m[ServerDock：连接已断开]\x1b[0m\r\n')
  }

  socket.onerror = () => {
    if (ws !== socket) return
    status.value = 'disconnected'
    statusMessage.value = '连接失败'
    term?.write('\r\n\x1b[31m[ServerDock：无法连接终端]\x1b[0m\r\n')
  }
}

function reconnect() {
  const previousSocket = ws
  ws = null
  previousSocket?.close()
  term?.clear()
  status.value = 'connecting'
  statusMessage.value = '正在重新连接…'
  connectWebSocket()
}

function goBack() {
  if (window.opener) {
    window.close()
    return
  }
  router.push({ path: '/containers', query: { server: serverId.value } })
}

function handleResize() {
  fitAddon?.fit()
}

onMounted(() => {
  initTerminal()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  const socket = ws
  ws = null
  socket?.close()
  term?.dispose()
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="terminal-page">
    <header class="terminal-toolbar">
      <button class="back-button" type="button" aria-label="返回容器管理" @click="goBack">
        <ArrowLeft :size="17" aria-hidden="true" />
      </button>
      <div class="terminal-context">
        <strong>{{ containerName || `服务器 ${serverId}` }}</strong>
        <span :class="`is-${status}`" aria-live="polite">
          <i aria-hidden="true" />{{ statusMessage }}
        </span>
      </div>
      <button v-if="status === 'disconnected'" class="reconnect-button" type="button" @click="reconnect">
        <RefreshCw :size="14" aria-hidden="true" />重新连接
      </button>
      <span v-else class="toolbar-end" aria-hidden="true" />
    </header>

    <div v-if="initError" class="terminal-error" role="alert">{{ initError }}</div>
    <main class="terminal-canvas" :aria-label="containerName ? `${containerName} 容器终端` : '服务器终端'">
      <div ref="terminalEl" class="terminal-inner" />
    </main>
  </div>
</template>

<style scoped>
.terminal-page {
  height: 100vh;
  height: 100dvh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--terminal);
}

.terminal-toolbar {
  min-height: 50px;
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background: #1d1e22;
  color: #fff;
}

.terminal-context {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.terminal-context strong {
  overflow: hidden;
  color: #f5f5f7;
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.terminal-context span {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 5px;
  color: #a1a1a6;
  font-size: 10px;
}

.terminal-context i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #ffbd45;
}

.terminal-context .is-connected i {
  background: #4dcc77;
}

.terminal-context .is-disconnected i {
  background: #ff6961;
}

.toolbar-end {
  width: 34px;
}

.back-button {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  background: transparent;
  color: #a1a1a6;
  cursor: pointer;
}

.back-button:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.reconnect-button {
  min-height: 30px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  margin-left: 4px;
  padding: 0 9px;
  border: 1px solid rgba(102, 168, 255, 0.34);
  border-radius: 7px;
  background: rgba(0, 113, 227, 0.14);
  color: #86bbff;
  cursor: pointer;
  font-size: 11px;
  font-weight: 650;
}

.reconnect-button:hover {
  background: rgba(0, 113, 227, 0.24);
}

.terminal-error {
  padding: 8px 14px;
  border-bottom: 1px solid rgba(255, 102, 92, 0.3);
  background: rgba(180, 35, 24, 0.15);
  color: #ff9c96;
  font-size: 12px;
  text-align: center;
}

.terminal-canvas {
  min-height: 0;
  flex: 1;
  padding: 10px 8px 8px 12px;
  overflow: hidden;
}

.terminal-inner {
  width: 100%;
  height: 100%;
}

@media (max-width: 660px) {
  .terminal-toolbar {
    min-height: calc(54px + env(safe-area-inset-top));
    gap: 8px;
    padding: max(6px, env(safe-area-inset-top)) 9px 6px;
  }

  .terminal-context {
    align-items: flex-start;
    justify-content: center;
    flex-direction: column;
    gap: 1px;
  }

  .terminal-context strong {
    max-width: 100%;
    font-size: 11px;
  }

  .terminal-context span {
    font-size: 9px;
  }

  .reconnect-button {
    min-height: 34px;
  }

  .terminal-canvas {
    padding: 7px 4px 4px 7px;
  }
}
</style>
