<script setup lang="ts">
defineProps<{
  collapsed?: boolean
}>()
</script>

<template>
  <div
    class="flex items-center gap-2.5 px-2 py-1.5"
    :class="collapsed ? 'justify-center' : ''"
  >
    <div class="size-7 rounded-md bg-primary flex items-center justify-center shrink-0">
      <UIcon name="i-lucide-zap" class="size-4 text-white" />
    </div>
    <span v-if="!collapsed" class="font-semibold text-sm text-highlighted truncate">
      Simplix CRM
    </span>
  </div>
</template>
