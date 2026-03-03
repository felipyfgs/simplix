<script setup lang="ts">
interface InboxSettings {
  phone_number: string
  webhook_verify_token: string
}

interface Inbox {
  id: string
  name: string
  channel_type: string
  settings?: InboxSettings
  created_at: string
}

defineProps<{ inboxes: Inbox[] }>()
const emit = defineEmits<{
  delete: [id: string]
  edit: [inbox: Inbox]
}>()

const config = useRuntimeConfig()

const channelIcon: Record<string, string> = {
  whatsapp: 'i-lucide-message-circle',
  quepasa: 'i-lucide-message-circle-code',
  email: 'i-lucide-mail',
  manual: 'i-lucide-inbox',
  phone: 'i-lucide-phone',
  web: 'i-lucide-globe'
}

type BadgeColor = 'success' | 'info' | 'neutral' | 'primary' | 'warning' | 'error' | 'secondary'

const channelColor: Record<string, BadgeColor> = {
  whatsapp: 'success',
  quepasa: 'success',
  email: 'info',
  manual: 'neutral',
  phone: 'primary',
  web: 'warning'
}

function webhookUrl(inbox: Inbox) {
  if (inbox.channel_type === 'quepasa') {
    return `${config.public.apiBase}/webhook/quepasa/${inbox.id}`
  }
  return `${config.public.apiBase}/webhook/whatsapp/${inbox.id}`
}

const toast = useToast()

async function copyWebhook(inbox: Inbox) {
  await navigator.clipboard.writeText(webhookUrl(inbox))
  toast.add({ title: 'URL copiada', icon: 'i-lucide-check', color: 'success' })
}
</script>

<template>
  <ul role="list" class="divide-y divide-default">
    <li
      v-for="inbox in inboxes"
      :key="inbox.id"
      class="flex items-start justify-between gap-3 py-4 px-4 sm:px-6"
    >
      <div class="flex items-start gap-3 min-w-0 flex-1">
        <UIcon
          :name="channelIcon[inbox.channel_type] ?? 'i-lucide-inbox'"
          class="size-5 mt-0.5 shrink-0 text-muted"
        />
        <div class="min-w-0">
          <p class="font-medium text-highlighted truncate">{{ inbox.name }}</p>
          <div class="flex items-center gap-2 mt-1">
            <UBadge
              :label="inbox.channel_type"
              :color="channelColor[inbox.channel_type] ?? 'neutral'"
              variant="subtle"
              size="xs"
            />
            <span v-if="inbox.settings?.phone_number" class="text-xs text-muted font-mono">
              {{ inbox.settings.phone_number }}
            </span>
          </div>
          <button
            v-if="inbox.channel_type === 'whatsapp' || inbox.channel_type === 'quepasa'"
            class="mt-1.5 flex items-center gap-1 text-xs text-muted hover:text-highlighted transition-colors font-mono truncate max-w-xs"
            @click="copyWebhook(inbox)"
          >
            <UIcon name="i-lucide-copy" class="size-3 shrink-0" />
            <span class="truncate">{{ webhookUrl(inbox) }}</span>
          </button>
        </div>
      </div>

      <div class="flex items-center gap-1 shrink-0">
        <UTooltip text="Configurações">
          <UButton
            :to="`/settings/inboxes/${inbox.id}`"
            icon="i-lucide-settings"
            size="xs"
            color="neutral"
            variant="ghost"
          />
        </UTooltip>
        <UButton
          icon="i-lucide-trash-2"
          size="xs"
          color="error"
          variant="ghost"
          @click="emit('delete', inbox.id)"
        />
      </div>
    </li>

    <li v-if="inboxes.length === 0" class="py-8 text-center text-sm text-muted">
      Nenhuma caixa de entrada configurada.
    </li>
  </ul>
</template>
