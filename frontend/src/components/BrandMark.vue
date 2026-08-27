<script setup>
defineProps({
  size: { type: Number, default: 32 },
  dark: Boolean
})
</script>

<template>
  <span
    class="brand-mark"
    :class="{ 'brand-mark-dark': dark }"
    :style="{ width: `${size}px`, height: `${size}px` }"
    aria-hidden="true"
  >
    <svg viewBox="0 0 24 24" fill="none">
      <path d="M5.25 5.5h10.5a2 2 0 0 1 2 2v1.25h-12.5V5.5Z" />
      <path d="M5.25 12.25h12.5v1.25a2 2 0 0 1-2 2H5.25v-3.25Z" />
      <path d="M3 3.25v13.5A2.25 2.25 0 0 0 5.25 19h11.5" />
      <path d="M20.25 8.75v8.5" />
      <circle cx="8" cy="7.15" r=".65" fill="currentColor" stroke="none" />
      <circle cx="8" cy="13.85" r=".65" fill="currentColor" stroke="none" />
    </svg>
  </span>
</template>

<style scoped>
.brand-mark {
  display: inline-grid;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 9px;
  background: var(--blue);
  color: #fff;
}

.brand-mark-dark {
  background: var(--ink);
}

svg {
  width: 68%;
  height: 68%;
  stroke: currentColor;
  stroke-width: 1.65;
  stroke-linecap: round;
  stroke-linejoin: round;
}
</style>
