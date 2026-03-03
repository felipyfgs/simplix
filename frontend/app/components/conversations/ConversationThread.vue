<script setup lang="ts">
interface Contact {
  id: string
  name: string
  avatar_url: string | null
}

interface Conv {
  id: string
  subject: string | null
  status: string
  contact_id: string
  inbox_id: string
  contact?: Contact
}

interface Msg {
  id: string
  sender_type: string
  content: string
  content_type: string
  created_at: string
}

const props = defineProps<{ conversation: Conv }>()
const emit = defineEmits<{ close: [] }>()

const api = useApi()
const toast = useToast()

const { data: msgs, refresh: refreshMsgs } = useAsyncData(
  () => `msgs-${props.conversation.id}`,
  () => api.get<Msg[]>(`/api/conversations/${props.conversation.id}/messages`).catch(() => []),
  { watch: [() => props.conversation.id], server: false }
)

const { data: convLive, refresh: refreshConv } = useAsyncData(
  () => `conv-live-${props.conversation.id}`,
  () => api.get<Conv>(`/api/conversations/${props.conversation.id}`),
  { watch: [() => props.conversation.id], server: false }
)

const conv = computed(() => convLive.value ?? props.conversation)

// Fetch inbox to check channel type
const { data: inbox } = useAsyncData(
  () => `inbox-type-${conv.value.inbox_id}`,
  () => conv.value.inbox_id
    ? api.get<{ id: string; channel_type: string }>(`/api/inboxes/${conv.value.inbox_id}`).catch(() => null)
    : Promise.resolve(null),
  { watch: [() => conv.value.inbox_id], server: false }
)

const isWhatsApp = computed(() => inbox.value?.channel_type === 'whatsapp')
const showTemplateModal = ref(false)

const newMsg = ref('')
const sending = ref(false)
const msgType = ref<'text' | 'note'>('text')

async function send() {
  if (!newMsg.value.trim()) return
  sending.value = true
  try {
    await api.post(`/api/conversations/${props.conversation.id}/messages`, {
      content: newMsg.value,
      content_type: msgType.value
    })
    newMsg.value = ''
    await refreshMsgs()
  }
  catch { toast.add({ title: 'Erro ao enviar mensagem', color: 'error' }) }
  finally { sending.value = false }
}

async function resolve() {
  await api.patch(`/api/conversations/${props.conversation.id}`, { status: 'resolved' })
  refreshConv()
  toast.add({ title: 'Conversa resolvida', icon: 'i-lucide-check', color: 'success' })
}

async function reopen() {
  await api.patch(`/api/conversations/${props.conversation.id}`, { status: 'open' })
  refreshConv()
}

const statusColor = { open: 'success', pending: 'warning', resolved: 'neutral', snoozed: 'info' } as const

const dropdownItems = [[{
  label: 'Resolver conversa',
  icon: 'i-lucide-check',
  onSelect: resolve
}, {
  label: 'Reabrir conversa',
  icon: 'i-lucide-rotate-ccw',
  onSelect: reopen
}]]

// Map our message to UChatMessage props
function msgRole(msg: Msg): 'user' | 'assistant' {
  return msg.sender_type === 'agent' ? 'user' : 'assistant'
}

function msgSide(msg: Msg): 'left' | 'right' {
  return msg.sender_type === 'agent' ? 'right' : 'left'
}

function msgVariant(msg: Msg): 'soft' | 'naked' | 'subtle' {
  if (msg.content_type === 'note') return 'subtle'
  return msg.sender_type === 'agent' ? 'soft' : 'naked'
}
</script>

<template>
  <UDashboardPanel id="conv-thread">
    <UDashboardNavbar :toggle="false">
      <template #leading>
        <UButton
          icon="i-lucide-arrow-left"
          color="neutral"
          variant="ghost"
          class="-ms-1.5"
          @click="emit('close')"
        />
        <div class="flex items-center gap-2 min-w-0">
          <div
            v-if="conv.contact?.avatar_url"
            class="size-7 rounded-full overflow-hidden shrink-0"
          >
            <img :src="conv.contact.avatar_url" :alt="conv.contact.name" class="w-full h-full object-cover" />
          </div>
          <div
            v-else
            class="size-7 rounded-full bg-primary/15 text-primary flex items-center justify-center text-xs font-semibold shrink-0"
          >
            {{ conv.contact ? conv.contact.name.split(' ').map((w: string) => w[0]).join('').slice(0, 2).toUpperCase() : '?' }}
          </div>
          <div class="min-w-0">
            <div class="flex items-center gap-1.5">
              <p class="text-sm font-semibold text-highlighted truncate leading-tight">
                {{ conv.contact?.name ?? 'Contato desconhecido' }}
              </p>
              <UIcon v-if="isWhatsApp" name="i-lucide-message-circle" class="size-3.5 text-success shrink-0" />
            </div>
            <p v-if="conv.subject" class="text-xs text-muted truncate leading-tight">
              {{ conv.subject }}
            </p>
          </div>
        </div>
      </template>

      <template #right>
        <UBadge
          :color="statusColor[conv.status as keyof typeof statusColor] ?? 'neutral'"
          :label="conv.status"
          variant="subtle"
          size="sm"
        />
        <UDropdownMenu :items="dropdownItems">
          <UButton icon="i-lucide-ellipsis-vertical" color="neutral" variant="ghost" />
        </UDropdownMenu>
      </template>
    </UDashboardNavbar>

    <!-- Messages -->
    <UChatMessages
      :should-scroll-to-bottom="true"
      :should-auto-scroll="false"
      :ui="{ root: 'flex-1 px-4', viewport: 'px-0' }"
    >
      <template v-if="!msgs?.length">
        <div class="flex items-center justify-center py-16 text-sm text-dimmed">
          Nenhuma mensagem ainda.
        </div>
      </template>

      <UChatMessage
        v-for="msg in msgs"
        :key="msg.id"
        :id="msg.id"
        :content="msg.content"
        :role="msgRole(msg)"
        :side="msgSide(msg)"
        :variant="msgVariant(msg)"
      >
        <template #content>
          <div
            :class="[
              'text-sm',
              msg.content_type === 'note' ? 'rounded-lg px-3 py-2 bg-warning-100 dark:bg-warning-900/30 border border-warning-300 dark:border-warning-700' : ''
            ]"
          >
            <p v-if="msg.content_type === 'note'" class="text-xs font-semibold text-warning-700 dark:text-warning-400 mb-1">
              Nota interna
            </p>
            <p class="whitespace-pre-wrap">{{ msg.content }}</p>
            <p class="text-xs opacity-50 mt-1 text-right">
              {{ new Date(msg.created_at).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' }) }}
            </p>
          </div>
        </template>
      </UChatMessage>
    </UChatMessages>

    <!-- Reply box -->
    <div class="pb-4 px-4 sm:px-6 shrink-0">
      <UChatPrompt
        v-model="newMsg"
        variant="subtle"
        :placeholder="msgType === 'note' ? 'Nota interna...' : 'Escrever mensagem...'"
        :loading="sending"
        :rows="2"
        @submit="send"
      >
        <template #header>
          <div class="flex items-center justify-between w-full px-1 pt-1">
            <div class="flex gap-1">
              <UButton
                :variant="msgType === 'text' ? 'soft' : 'ghost'"
                size="xs"
                label="Mensagem"
                icon="i-lucide-reply"
                @click="msgType = 'text'"
              />
              <UButton
                :variant="msgType === 'note' ? 'soft' : 'ghost'"
                color="warning"
                size="xs"
                label="Nota interna"
                icon="i-lucide-pencil"
                @click="msgType = 'note'"
              />
            </div>
            <UButton
              v-if="isWhatsApp"
              size="xs"
              color="success"
              variant="ghost"
              icon="i-lucide-layout-template"
              label="Template"
              @click="showTemplateModal = true"
            />
          </div>
        </template>

        <template #trailing>
          <UChatPromptSubmit color="neutral" />
        </template>
      </UChatPrompt>
    </div>
  </UDashboardPanel>

  <ConversationsWhatsAppTemplateModal
    v-if="showTemplateModal && isWhatsApp && conv.inbox_id"
    :inbox-id="conv.inbox_id"
    :conversation-id="conv.id"
    @close="showTemplateModal = false; refreshMsgs()"
  />
</template>
