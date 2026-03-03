<script setup lang="ts">
const api = useApi()
const route = useRoute()

const page = ref(1)
const status = ref('open')
const q = ref('')

const { data, pending, refresh } = useAsyncData('conv-inbox-list', () =>
  api.get<{ data: Conv[]; pagination: any }>('/api/conversations', {
    page: String(page.value),
    limit: '30',
    status: status.value,
    q: q.value
  }).catch(() => ({ data: [], pagination: { total: 0 } })),
  { watch: [page, status, q], server: false }
)

const convs = computed(() => data.value?.data ?? [])
const selectedId = computed(() => route.params.id as string | undefined)

const statusTabs = [
  { label: 'Abertas', value: 'open' },
  { label: 'Pendentes', value: 'pending' },
  { label: 'Resolvidas', value: 'resolved' },
  { label: 'Todas', value: '' }
]

const statusColor = {
  open: 'success', pending: 'warning', resolved: 'neutral', snoozed: 'info'
} as const

interface Conv {
  id: string
  subject: string | null
  status: string
  contact_id: string
  last_activity_at: string
}
</script>

<template>
  <div class="flex flex-col h-full overflow-hidden">
    <!-- Filters header -->
    <div class="p-3 border-b border-default space-y-2 flex-shrink-0">
      <UInput
        v-model="q"
        icon="i-lucide-search"
        placeholder="Buscar..."
        size="sm"
        class="w-full"
        @update:model-value="page = 1"
      />
      <div class="flex gap-1 flex-wrap">
        <UButton
          v-for="tab in statusTabs"
          :key="tab.value"
          :variant="status === tab.value ? 'soft' : 'ghost'"
          size="xs"
          :label="tab.label"
          @click="status = tab.value; page = 1"
        />
      </div>
    </div>

    <!-- Conversation list -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="pending" class="flex justify-center py-8">
        <UIcon name="i-lucide-loader-2" class="animate-spin text-muted size-5" />
      </div>

      <div v-else-if="convs.length === 0" class="text-center py-12 text-muted text-sm">
        Nenhuma conversa encontrada.
      </div>

      <NuxtLink
        v-for="conv in convs"
        :key="conv.id"
        :to="`/conversations/${conv.id}`"
        class="block border-b border-default p-3 hover:bg-elevated/60 transition-colors"
        :class="{ 'bg-primary/10 border-l-2 border-l-primary': conv.id === selectedId }"
      >
        <div class="flex items-start justify-between gap-2">
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium truncate">
              {{ conv.subject ?? 'Sem assunto' }}
            </p>
            <p class="text-xs text-muted mt-0.5">
              {{ new Date(conv.last_activity_at).toLocaleDateString('pt-BR') }}
            </p>
          </div>
          <UBadge
            :color="statusColor[conv.status as keyof typeof statusColor] ?? 'neutral'"
            variant="subtle"
            size="xs"
            :label="conv.status"
          />
        </div>
      </NuxtLink>
    </div>

    <!-- New conversation button -->
    <div class="p-3 border-t border-default flex-shrink-0">
      <UButton
        icon="i-lucide-plus"
        label="Nova Conversa"
        variant="soft"
        size="sm"
        class="w-full"
        to="/contacts"
      />
    </div>
  </div>
</template>
