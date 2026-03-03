<script setup lang="ts">
import { format, isToday, isYesterday } from 'date-fns'
import { ptBR } from 'date-fns/locale'

interface Contact {
  id: string
  name: string
  avatar_url: string | null
}

interface LastMessage {
  sender_type: string
  content: string
  content_type: string
}

interface Conv {
  id: string
  subject: string | null
  status: string
  contact_id: string
  last_activity_at: string
  contact?: Contact
  last_message?: LastMessage
}

const props = defineProps<{ conversations: Conv[] }>()
const selectedConv = defineModel<Conv | null>()

const convRefs = ref<Record<string, Element | null>>({})

watch(selectedConv, () => {
  if (!selectedConv.value) return
  const el = convRefs.value[selectedConv.value.id]
  if (el) el.scrollIntoView({ block: 'nearest' })
})

function initials(name: string) {
  return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()
}

function formatTime(iso: string) {
  const d = new Date(iso)
  if (isToday(d)) return format(d, 'HH:mm')
  if (isYesterday(d)) return 'Ontem'
  return format(d, 'dd/MM', { locale: ptBR })
}

function messagePreview(msg: LastMessage | undefined) {
  if (!msg) return 'Sem mensagens ainda'
  if (msg.content_type === 'note') return '📝 ' + msg.content
  if (msg.sender_type === 'agent') return '↗ ' + msg.content
  return msg.content
}

const statusColors: Record<string, string> = {
  open:     'bg-success-500',
  pending:  'bg-warning-500',
  resolved: 'bg-neutral-400',
  snoozed:  'bg-info-500'
}

defineShortcuts({
  arrowdown: () => {
    const index = props.conversations.findIndex(c => c.id === selectedConv.value?.id)
    if (index === -1) selectedConv.value = props.conversations[0] ?? null
    else if (index < props.conversations.length - 1) selectedConv.value = props.conversations[index + 1] ?? null
  },
  arrowup: () => {
    const index = props.conversations.findIndex(c => c.id === selectedConv.value?.id)
    if (index > 0) selectedConv.value = props.conversations[index - 1] ?? null
  }
})
</script>

<template>
  <div class="overflow-y-auto flex-1">
    <div v-if="conversations.length === 0" class="py-16 text-center text-sm text-dimmed">
      Nenhuma conversa encontrada.
    </div>

    <div
      v-for="conv in conversations"
      :key="conv.id"
      :ref="(el) => { convRefs[conv.id] = el as Element | null }"
      class="flex items-start gap-3 px-4 py-3 cursor-pointer border-b border-default transition-colors border-l-2"
      :class="[
        selectedConv?.id === conv.id
          ? 'border-l-primary bg-primary/8'
          : 'border-l-transparent hover:bg-elevated/60'
      ]"
      @click="selectedConv = conv"
    >
      <!-- Avatar -->
      <div class="shrink-0 mt-0.5">
        <div
          v-if="conv.contact?.avatar_url"
          class="size-8 rounded-full overflow-hidden"
        >
          <img :src="conv.contact.avatar_url" :alt="conv.contact.name" class="w-full h-full object-cover" />
        </div>
        <div
          v-else
          class="size-8 rounded-full bg-primary/15 text-primary flex items-center justify-center text-xs font-semibold"
        >
          {{ conv.contact ? initials(conv.contact.name) : '?' }}
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <!-- Row 1: name + timestamp -->
        <div class="flex items-center justify-between gap-2 mb-0.5">
          <p class="text-sm font-semibold text-highlighted truncate">
            {{ conv.contact?.name ?? 'Contato desconhecido' }}
          </p>
          <span class="text-xs text-dimmed shrink-0">
            {{ formatTime(conv.last_activity_at) }}
          </span>
        </div>

        <!-- Row 2: subject -->
        <p v-if="conv.subject" class="text-xs text-muted truncate mb-0.5">
          {{ conv.subject }}
        </p>

        <!-- Row 3: last message preview + status dot -->
        <div class="flex items-center gap-2">
          <p class="text-xs text-dimmed truncate flex-1">
            {{ messagePreview(conv.last_message) }}
          </p>
          <span
            :class="['size-2 rounded-full shrink-0', statusColors[conv.status] ?? 'bg-neutral-400']"
            :title="conv.status"
          />
        </div>
      </div>
    </div>
  </div>
</template>
