<script setup lang="ts">
defineProps<{
  label: string
  value: number | string
  icon: string
  sub?: string
}>()
</script>

<template>
  <div class="flex flex-col gap-1 p-4 border border-default rounded-lg bg-default">
    <div class="flex items-center gap-2 text-muted">
      <UIcon :name="icon" class="size-4 shrink-0" />
      <span class="text-xs font-medium uppercase tracking-wide">{{ label }}</span>
    </div>
    <p class="text-3xl font-semibold text-highlighted mt-1">{{ value }}</p>
    <p v-if="sub" class="text-xs text-muted">{{ sub }}</p>
  </div>
</template>
