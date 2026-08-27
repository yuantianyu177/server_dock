<script setup>
defineProps({
  size: { type: Number, default: 32 }
})
</script>

<template>
  <img
    class="brand-mark"
    :style="{ width: `${size}px`, height: `${size}px` }"
    src="/serverdock-icon.svg"
    alt=""
    aria-hidden="true"
    draggable="false"
  />
</template>

<style scoped>
.brand-mark {
  display: block;
  flex: 0 0 auto;
  object-fit: contain;
  user-select: none;
}
</style>
