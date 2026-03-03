<script setup lang="ts">
import { breakpointsTailwind } from '@vueuse/core'

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
  inbox_id: string
  last_activity_at: string
  contact?: Contact
  last_message?: LastMessage
}

const tabItems = [
  { label: 'Abertas',   value: 'open' },
  { label: 'Pendentes', value: 'pending' },
  { label: 'Resolvidas', value: 'resolved' },
  { label: 'Todas',     value: '' }
]
const selectedTab = ref('open')

const api = useApi()
const q = ref('')

// Filters
const showFilters = ref(false)
const filterInbox = ref('')
const filterLabel = ref('')
const filterAssignedTo = ref('')

interface Inbox { id: string; name: string }
interface Label { id: string; name: string; color: string }
interface User  { id: string; name: string }

const { data: inboxes } = useAsyncData('filter-inboxes', () =>
  api.get<Inbox[]>('/api/inboxes').catch(() => [] as Inbox[])
)
const { data: labels } = useAsyncData('filter-labels', () =>
  api.get<Label[]>('/api/labels').catch(() => [] as Label[])
)
const { data: users } = useAsyncData('filter-users', () =>
  api.get<User[]>('/api/users').catch(() => [] as User[])
)

const hasActiveFilters = computed(() => !!(filterInbox.value || filterLabel.value || filterAssignedTo.value))

function clearFilters() {
  filterInbox.value = ''
  filterLabel.value = ''
  filterAssignedTo.value = ''
}

const { data, refresh } = useAsyncData('conversations-list', () => {
  const params: Record<string, string> = { limit: '50' }
  if (selectedTab.value) params.status = selectedTab.value
  if (q.value) params.q = q.value
  if (filterInbox.value) params.inbox_id = filterInbox.value
  if (filterLabel.value) params.label = filterLabel.value
  if (filterAssignedTo.value) params.assigned_to = filterAssignedTo.value
  return api.get<{ data: Conv[]; pagination: any }>('/api/conversations', params)
    .catch(() => ({ data: [], pagination: { total: 0 } }))
}, { watch: [selectedTab, q, filterInbox, filterLabel, filterAssignedTo], server: false })

const conversations = computed(() => data.value?.data ?? [])
const selectedConv = ref<Conv | null>(null)

const isConvPanelOpen = computed({
  get() { return !!selectedConv.value },
  set(value: boolean) { if (!value) selectedConv.value = null }
})

watch(conversations, () => {
  if (selectedConv.value && !conversations.value.find(c => c.id === selectedConv.value?.id)) {
    selectedConv.value = null
  }
})

const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('lg')

useIntervalFn(refresh, 15000)
</script>

<template>
  <!-- Column 1: conversation list -->
  <UDashboardPanel
    id="conv-list"
    :default-size="25"
    :min-size="20"
    :max-size="35"
    resizable
  >
    <UDashboardNavbar title="Conversas">
      <template #leading>
        <UDashboardSidebarCollapse />
      </template>
      <template #trailing>
        <UBadge :label="String(conversations.length)" variant="subtle" />
      </template>
    </UDashboardNavbar>

    <!-- Status tabs -->
    <div class="px-3 pt-2 pb-1 border-b border-default shrink-0">
      <UTabs
        v-model="selectedTab"
        :items="tabItems"
        :content="false"
        size="xs"
        class="w-full"
        :ui="{ list: 'w-full' }"
      />
    </div>

    <!-- Search + filter toggle -->
    <div class="px-3 py-2 shrink-0 flex gap-2">
      <UInput
        v-model="q"
        icon="i-lucide-search"
        placeholder="Buscar conversas..."
        size="sm"
        class="flex-1"
      />
      <UButton
        icon="i-lucide-sliders-horizontal"
        color="neutral"
        :variant="hasActiveFilters ? 'solid' : 'ghost'"
        size="sm"
        @click="showFilters = !showFilters"
      />
    </div>

    <!-- Expanded filters -->
    <div v-if="showFilters" class="px-3 pb-2 space-y-2 shrink-0 border-b border-default">
      <USelect
        v-model="filterInbox"
        :options="[{ label: 'Todas as inboxes', value: '' }, ...(inboxes ?? []).map(i => ({ label: i.name, value: i.id }))]"
        size="sm"
        class="w-full"
        placeholder="Inbox"
      />
      <USelect
        v-model="filterLabel"
        :options="[{ label: 'Todas as etiquetas', value: '' }, ...(labels ?? []).map(l => ({ label: l.name, value: l.name }))]"
        size="sm"
        class="w-full"
        placeholder="Etiqueta"
      />
      <USelect
        v-model="filterAssignedTo"
        :options="[{ label: 'Todos os agentes', value: '' }, ...(users ?? []).map(u => ({ label: u.name, value: u.id }))]"
        size="sm"
        class="w-full"
        placeholder="Atribuído a"
      />
      <UButton v-if="hasActiveFilters" label="Limpar filtros" color="neutral" variant="ghost" size="xs" @click="clearFilters" />
    </div>

    <ConversationsConversationList
      v-model="selectedConv"
      :conversations="conversations"
    />
  </UDashboardPanel>

  <!-- Column 2: thread or empty state -->
  <ConversationsConversationThread
    v-if="selectedConv && !isMobile"
    :conversation="selectedConv"
    @close="selectedConv = null"
  />
  <div v-else-if="!isMobile" class="hidden lg:flex flex-1 items-center justify-center flex-col gap-3">
    <UIcon name="i-lucide-message-square" class="size-16 text-dimmed" />
    <p class="text-sm text-muted">Selecione uma conversa</p>
  </div>

  <!-- Column 3: side panel -->
  <ConversationsConversationSidePanel
    v-if="selectedConv && !isMobile"
    :conversation="selectedConv"
  />

  <!-- Mobile: slideover -->
  <ClientOnly>
    <USlideover v-if="isMobile" v-model:open="isConvPanelOpen">
      <template #content>
        <ConversationsConversationThread
          v-if="selectedConv"
          :conversation="selectedConv"
          @close="selectedConv = null"
        />
      </template>
    </USlideover>
  </ClientOnly>
</template>
