<script setup lang="ts">
import type { FilterCondition } from '~/components/contacts/FilterPanel.vue'

interface Contact {
  id: string
  name: string
  email: string | null
  phone: string | null
  company: string | null
  avatar_url: string | null
  status: string
  score: number
  custom_attributes: Record<string, string> | null
}

interface Pagination {
  page: number
  limit: number
  total: number
}

const api = useApi()

const page = ref(1)
const q = ref('')
const showFilterPanel = ref(false)
const activeFilters = ref<FilterCondition[]>([])

const filterParams = computed(() =>
  activeFilters.value
    .filter(c => c.op === 'present' || c.op === 'not_present' || c.value)
    .map(c => (c.op === 'present' || c.op === 'not_present')
      ? `${c.field}:${c.op}`
      : `${c.field}:${c.op}:${c.value}`)
)

const hasActiveFilters = computed(() => filterParams.value.length > 0)

const { data, refresh, pending } = await useAsyncData('contacts', () =>
  api.get<{ data: Contact[]; pagination: Pagination }>('/api/contacts', {
    page: String(page.value),
    limit: '20',
    q: q.value,
    filter: filterParams.value
  }).catch(() => ({ data: [], pagination: { page: 1, limit: 20, total: 0 } })),
  { watch: [page, q, filterParams] }
)

const contacts = computed(() => data.value?.data ?? [])
const total = computed(() => data.value?.pagination.total ?? 0)

function applyFilters(conditions: FilterCondition[]) {
  activeFilters.value = conditions
  showFilterPanel.value = false
  page.value = 1
}

function clearFilters() {
  activeFilters.value = []
  page.value = 1
}
</script>

<template>
  <UDashboardPanel id="contacts">
    <template #header>
      <UDashboardNavbar title="Contatos">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>

        <template #right>
          <UInput
            v-model="q"
            icon="i-lucide-search"
            placeholder="Pesquisar..."
            class="w-44"
            :ui="{ base: 'bg-elevated/60' }"
            @update:model-value="page = 1"
          />

          <UPopover
            v-model:open="showFilterPanel"
            :content="{ align: 'end', side: 'bottom', sideOffset: 8 }"
          >
            <div class="relative">
              <UButton
                icon="i-lucide-list-filter"
                color="neutral"
                variant="ghost"
                :class="hasActiveFilters ? 'text-primary' : ''"
              />
              <span
                v-if="hasActiveFilters"
                class="absolute -top-0.5 -right-0.5 size-2 rounded-full bg-primary pointer-events-none"
              />
            </div>
            <template #content>
              <ContactsFilterPanel
                @apply="applyFilters"
                @clear="clearFilters"
                @close="showFilterPanel = false"
              />
            </template>
          </UPopover>

          <USeparator orientation="vertical" class="h-4" />

          <ContactsContactCreateModal @created="refresh" />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <!-- Active filter badges -->
      <div v-if="hasActiveFilters" class="flex flex-wrap items-center gap-1.5 pb-3 border-b border-default">
        <UBadge
          v-for="(f, i) in activeFilters"
          :key="i"
          color="primary"
          variant="subtle"
          class="cursor-pointer gap-1"
          @click="activeFilters.splice(i, 1); page = 1"
        >
          <span>{{ f.field }}: {{ f.op }}<template v-if="f.value"> "{{ f.value }}"</template></span>
          <UIcon name="i-lucide-x" class="size-3" />
        </UBadge>
        <UButton label="Limpar todos" variant="link" color="neutral" size="xs" @click="clearFilters" />
      </div>

      <!-- Loading -->
      <div v-if="pending" class="flex justify-center py-12">
        <UIcon name="i-lucide-loader-circle" class="size-6 text-muted animate-spin" />
      </div>

      <!-- List -->
      <div v-else class="flex flex-col gap-2">
        <ContactsContactCard
          v-for="c in contacts"
          :key="c.id"
          :contact="c"
          @deleted="refresh"
          @updated="refresh"
        />
        <p v-if="contacts.length === 0" class="text-center text-muted py-12 text-sm">
          Nenhum contato encontrado.
        </p>
      </div>

      <!-- Pagination -->
      <div class="flex items-center justify-between gap-3 border-t border-default pt-4 mt-auto">
        <p class="text-sm text-muted">
          {{ total.toLocaleString('pt-BR') }} contato(s)
        </p>
        <UPagination
          :default-page="page"
          :items-per-page="20"
          :total="total"
          @update:page="(p: number) => page = p"
        />
      </div>
    </template>
  </UDashboardPanel>
</template>
