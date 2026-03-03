<script setup lang="ts">
interface Webhook {
  id: string
  url: string
  subscriptions: string[]
  enabled: boolean
}

defineProps<{ webhooks: Webhook[] }>()
const emit = defineEmits<{ delete: [id: string] }>()
</script>

<template>
  <ul role="list" class="divide-y divide-default">
    <li
      v-for="hook in webhooks"
      :key="hook.id"
      class="flex items-start justify-between gap-3 py-4 px-4 sm:px-6"
    >
      <div class="min-w-0 flex-1">
        <p class="font-mono text-xs text-highlighted truncate">{{ hook.url }}</p>
        <div class="flex flex-wrap gap-1 mt-1.5">
          <UBadge
            v-for="evt in hook.subscriptions"
            :key="evt"
            :label="evt"
            variant="subtle"
            size="xs"
          />
        </div>
      </div>

      <div class="flex items-center gap-2 shrink-0">
        <UIcon
          :name="hook.enabled ? 'i-lucide-check-circle' : 'i-lucide-x-circle'"
          :class="hook.enabled ? 'text-success' : 'text-error'"
          class="size-4"
        />
        <UButton
          icon="i-lucide-trash-2"
          size="xs"
          color="error"
          variant="ghost"
          @click="emit('delete', hook.id)"
        />
      </div>
    </li>
    <li v-if="webhooks.length === 0" class="py-8 text-center text-sm text-muted">
      Nenhum webhook configurado.
    </li>
  </ul>
</template>
