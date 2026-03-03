<script setup lang="ts">
interface Company {
  id: string
  name: string
  domain: string | null
  phone: string | null
  website: string | null
  industry: string | null
  description: string | null
  contacts_count: number
  updated_at: string
}

interface Pagination { page: number; limit: number; total: number }

const api = useApi()

const page = ref(1)
const q = ref('')
const sort = ref('name')

const sortOptions = [
  { label: 'Nome (A-Z)', value: 'name' },
  { label: 'Nome (Z-A)', value: '-name' },
  { label: 'Domínio (A-Z)', value: 'domain' },
  { label: 'Domínio (Z-A)', value: '-domain' },
  { label: 'Mais recentes', value: '-created_at' },
  { label: 'Mais antigos', value: 'created_at' },
]

let debounceTimer: ReturnType<typeof setTimeout> | null = null
const debouncedQ = ref('')

watch(q, (val) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    debouncedQ.value = val
    page.value = 1
  }, 300)
})

watch(sort, () => { page.value = 1 })

const { data, pending } = await useAsyncData('companies', () => {
  const params: Record<string, string> = {
    page: String(page.value),
    limit: '25',
    sort: sort.value,
  }
  if (debouncedQ.value) params.q = debouncedQ.value
  return api.get<{ data: Company[]; pagination: Pagination }>('/api/companies', params)
    .catch(() => ({ data: [] as Company[], pagination: { page: 1, limit: 25, total: 0 } }))
}, { watch: [page, debouncedQ, sort] })

const companies = computed(() => data.value?.data ?? [])
const total = computed(() => data.value?.pagination.total ?? 0)
</script>

<template>
  <UDashboardPanel id="companies">
    <template #header>
      <UDashboardNavbar title="Empresas">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
        <template #right>
          <USelect
            v-model="sort"
            :items="sortOptions"
            size="sm"
            class="w-44"
          />
          <UInput
            v-model="q"
            icon="i-lucide-search"
            placeholder="Buscar empresas..."
            size="sm"
            class="w-52"
            :ui="{ base: 'bg-elevated/60' }"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div v-if="pending" class="flex justify-center py-16">
        <UIcon name="i-lucide-loader-circle" class="size-6 text-muted animate-spin" />
      </div>

      <div v-else-if="companies.length === 0" class="flex flex-col items-center justify-center py-16 gap-3">
        <UIcon name="i-lucide-building-2" class="size-10 text-muted" />
        <p class="text-sm text-muted">Nenhuma empresa encontrada.</p>
      </div>

      <div v-else class="flex flex-col gap-2">
        <CompaniesCompanyCard
          v-for="c in companies"
          :key="c.id"
          :company="c"
        />
      </div>

      <div class="flex items-center justify-between gap-3 border-t border-default pt-4 mt-auto">
        <p class="text-sm text-muted">
          {{ total.toLocaleString('pt-BR') }} empresa(s)
        </p>
        <UPagination
          :default-page="page"
          :items-per-page="25"
          :total="total"
          @update:page="(p: number) => page = p"
        />
      </div>
    </template>
  </UDashboardPanel>
</template>
