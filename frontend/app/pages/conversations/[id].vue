<script setup lang="ts">
navigateTo('/conversations', { replace: true })

const route = useRoute()
const api = useApi()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: conv, refresh: refreshConv } = useAsyncData(
  () => `conv-${id.value}`,
  () => api.get<Conv>(`/api/conversations/${id.value}`),
  { watch: [id], server: false }
)

const { data: msgs, refresh: refreshMsgs } = useAsyncData(
  () => `msgs-${id.value}`,
  () => api.get<Msg[]>(`/api/conversations/${id.value}/messages`).catch(() => []),
  { watch: [id], server: false }
)

const newMsg = ref('')
const sending = ref(false)
const msgType = ref<'text' | 'note'>('text')
const msgArea = ref<HTMLElement | null>(null)

async function send() {
  if (!newMsg.value.trim()) return
  sending.value = true
  try {
    await api.post(`/api/conversations/${id.value}/messages`, {
      content: newMsg.value,
      content_type: msgType.value
    })
    newMsg.value = ''
    await refreshMsgs()
    nextTick(() => {
      if (msgArea.value) msgArea.value.scrollTop = msgArea.value.scrollHeight
    })
  }
  catch { toast.add({ title: 'Erro ao enviar mensagem', color: 'error' }) }
  finally { sending.value = false }
}

async function resolve() {
  await api.patch(`/api/conversations/${id.value}`, { status: 'resolved' })
  refreshConv()
  toast.add({ title: 'Conversa resolvida', color: 'success' })
}

async function reopen() {
  await api.patch(`/api/conversations/${id.value}`, { status: 'open' })
  refreshConv()
}

const statusColor = { open: 'success', pending: 'warning', resolved: 'neutral', snoozed: 'info' } as const

interface Conv { id: string; subject: string | null; status: string; contact_id: string }
interface Msg { id: string; sender_type: string; content: string; content_type: string; created_at: string }
</script>

<template>
  <div class="flex flex-col h-full overflow-hidden">
    <!-- Conversation header -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-default flex-shrink-0">
      <div class="flex items-center gap-3 min-w-0">
        <p class="font-semibold truncate">{{ conv?.subject ?? 'Conversa sem assunto' }}</p>
        <UBadge
          v-if="conv"
          :color="statusColor[conv.status as keyof typeof statusColor] ?? 'neutral'"
          :label="conv.status"
          size="sm"
          variant="subtle"
        />
      </div>
      <div class="flex items-center gap-2 flex-shrink-0">
        <UButton
          v-if="conv?.status === 'open'"
          icon="i-lucide-check"
          size="xs"
          variant="soft"
          color="success"
          label="Resolver"
          @click="resolve"
        />
        <UButton
          v-else-if="conv"
          icon="i-lucide-rotate-ccw"
          size="xs"
          variant="soft"
          label="Reabrir"
          @click="reopen"
        />
      </div>
    </div>

    <!-- Messages area -->
    <div ref="msgArea" class="flex-1 overflow-y-auto p-4 space-y-3">
      <div v-if="!msgs?.length" class="text-center py-12 text-muted text-sm">
        Nenhuma mensagem ainda. Inicie a conversa abaixo.
      </div>

      <div
        v-for="msg in msgs"
        :key="msg.id"
        :class="['flex', msg.sender_type === 'agent' ? 'justify-end' : 'justify-start']"
      >
        <div
          :class="[
            'max-w-sm rounded-xl px-4 py-2 text-sm shadow-sm',
            msg.content_type === 'note'
              ? 'bg-warning-50 dark:bg-warning-900/20 border border-warning-200 dark:border-warning-800 text-warning-800 dark:text-warning-200'
              : msg.sender_type === 'agent'
                ? 'bg-primary text-white'
                : 'bg-elevated border border-default'
          ]"
        >
          <p v-if="msg.content_type === 'note'" class="text-xs font-semibold mb-1 opacity-70">
            Nota interna
          </p>
          <p class="whitespace-pre-wrap">{{ msg.content }}</p>
          <p class="text-xs opacity-60 mt-1 text-right">
            {{ new Date(msg.created_at).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' }) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Send message -->
    <div class="border-t border-default p-3 space-y-2 flex-shrink-0">
      <div class="flex gap-1">
        <UButton
          :variant="msgType === 'text' ? 'soft' : 'ghost'"
          size="xs"
          label="Mensagem"
          @click="msgType = 'text'"
        />
        <UButton
          :variant="msgType === 'note' ? 'soft' : 'ghost'"
          color="warning"
          size="xs"
          label="Nota interna"
          @click="msgType = 'note'"
        />
      </div>
      <div class="flex gap-2">
        <UTextarea
          v-model="newMsg"
          :placeholder="msgType === 'note' ? 'Nota interna...' : 'Escrever mensagem...'"
          :rows="2"
          class="flex-1 text-sm"
          @keydown.ctrl.enter="send"
        />
        <UButton
          icon="i-lucide-send"
          :loading="sending"
          size="sm"
          @click="send"
        />
      </div>
      <p class="text-xs text-muted">Ctrl+Enter para enviar</p>
    </div>
  </div>
</template>
