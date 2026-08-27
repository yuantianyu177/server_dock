<script setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const serverId = computed(() => route.params.serverId)
const containerName = computed(() => route.query.container || null)

const terminalEl = ref(null)
const status = ref('connecting')  // connecting | connected | disconnected
const statusMsg = ref('Connecting…')

let term = null
let fitAddon = null
let ws = null

async function initTerminal() {
  const { Terminal } = await import('@xterm/xterm')
  const { FitAddon } = await import('@xterm/addon-fit')
  await import('@xterm/xterm/css/xterm.css')

  term = new Terminal({
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    fontSize: 13,
    lineHeight: 1.4,
    cursorBlink: true,
    theme: {
      background: '#0F0F0E',
      foreground: '#E8E5DF',
      cursor: '#C96442',
      cursorAccent: '#0F0F0E',
      selectionBackground: 'rgba(201, 100, 66, 0.3)',
      black: '#1A1917',
      red: '#C0392B',
      green: '#2A7A4E',
      yellow: '#896B0A',
      blue: '#1A5F9E',
      magenta: '#8B5E9E',
      cyan: '#1A7D7D',
      white: '#D8D4CE',
      brightBlack: '#6B6762',
      brightRed: '#E74C3C',
      brightGreen: '#27AE60',
      brightYellow: '#F39C12',
      brightBlue: '#2980B9',
      brightMagenta: '#9B59B6',
      brightCyan: '#16A085',
      brightWhite: '#F5F3EE'
    }
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalEl.value)
  fitAddon.fit()

  connectWebSocket()

  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(data)
    }
  })

  term.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'resize', cols, rows }))
    }
  })
}

function connectWebSocket() {
  const token = localStorage.getItem('token') || ''
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const host = location.host

  let url
  if (containerName.value) {
    url = `${proto}://${host}/api/terminal/container/ws/${serverId.value}/${containerName.value}?token=${token}`
  } else {
    url = `${proto}://${host}/api/terminal/ws/${serverId.value}?token=${token}`
  }

  ws = new WebSocket(url)

  ws.onopen = () => {
    status.value = 'connected'
    statusMsg.value = containerName.value
      ? `${containerName.value} — connected`
      : 'Server terminal — connected'
    fitAddon?.fit()
  }

  ws.onmessage = (e) => {
    term?.write(e.data)
  }

  ws.onclose = () => {
    status.value = 'disconnected'
    statusMsg.value = 'Disconnected'
    term?.write('\r\n\x1b[31m[Connection closed]\x1b[0m\r\n')
  }

  ws.onerror = () => {
    status.value = 'disconnected'
    statusMsg.value = 'Connection error'
    term?.write('\r\n\x1b[31m[Connection error]\x1b[0m\r\n')
  }
}

function reconnect() {
  ws?.close()
  term?.clear()
  status.value = 'connecting'
  statusMsg.value = 'Reconnecting…'
  connectWebSocket()
}

function handleResize() {
  fitAddon?.fit()
}

onMounted(() => {
  initTerminal()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  ws?.close()
  term?.dispose()
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="terminal-page">
    <!-- Header bar -->
    <div class="terminal-header">
      <button class="term-back" @click="router.back()">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
        Back
      </button>

      <div class="term-info">
        <div class="term-status-dot" :class="status" />
        <span class="term-status-text">{{ statusMsg }}</span>
      </div>

      <button v-if="status === 'disconnected'" class="term-reconnect" @click="reconnect">
        Reconnect
      </button>
      <div v-else style="width:80px" />
    </div>

    <!-- Terminal area -->
    <div class="terminal-wrap">
      <div ref="terminalEl" class="terminal-inner" />
    </div>
  </div>
</template>

<style scoped>
.terminal-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #0F0F0E;
}

.terminal-header {
  height: 44px;
  background: #1A1917;
  border-bottom: 1px solid rgba(255,255,255,0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  flex-shrink: 0;
}

.term-back {
  display: flex;
  align-items: center;
  gap: 5px;
  color: #807B75;
  font-size: 13px;
  cursor: pointer;
  padding: 5px 8px;
  border-radius: 5px;
  transition: all 0.15s;
  background: none;
  border: none;
  font-family: var(--font-sans);
  width: 80px;
}

.term-back:hover {
  color: #E8E5DF;
  background: rgba(255,255,255,0.06);
}

.term-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.term-status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.term-status-dot.connecting { background: #F39C12; }
.term-status-dot.connected { background: #27AE60; }
.term-status-dot.disconnected { background: #E74C3C; }

.term-status-text {
  font-family: var(--font-mono);
  font-size: 12.5px;
  color: #9B9895;
}

.term-reconnect {
  font-size: 12.5px;
  color: #C96442;
  cursor: pointer;
  background: rgba(201, 100, 66, 0.12);
  border: 1px solid rgba(201, 100, 66, 0.25);
  border-radius: 5px;
  padding: 4px 10px;
  font-family: var(--font-sans);
  transition: all 0.15s;
  width: 80px;
  text-align: center;
}

.term-reconnect:hover {
  background: rgba(201, 100, 66, 0.2);
}

.terminal-wrap {
  flex: 1;
  overflow: hidden;
  padding: 8px;
}

.terminal-inner {
  width: 100%;
  height: 100%;
}
</style>
