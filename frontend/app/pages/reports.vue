<script setup lang="ts">
import { sub, format } from 'date-fns'
import type { Period, Range } from '~/types'

interface KV { key: string; value: number }
interface AgentStat { agent: string; total: number; resolved: number; open: number }
interface TimeSeriesPoint { timestamp: string; value: number }
interface TimeSeriesReport {
  conversations: TimeSeriesPoint[]
  incoming_messages: TimeSeriesPoint[]
  outgoing_messages: TimeSeriesPoint[]
  resolutions: TimeSeriesPoint[]
}

const api = useApi()

// ─── Period / Range filters ──────────────────────────────────────────────────
const range = shallowRef<Range>({
  start: sub(new Date(), { days: 29 }),
  end: new Date()
})
const period = ref<Period>('daily')

const groupBy = computed(() => {
  const map: Record<Period, string> = { daily: 'day', weekly: 'week', monthly: 'month' }
  return map[period.value]
})

const fromStr = computed(() => format(range.value.start, 'yyyy-MM-dd'))
const toStr = computed(() => format(range.value.end, 'yyyy-MM-dd'))

// ─── Active tab ──────────────────────────────────────────────────────────────
const tab = ref('overview')
const tabItems = [
  { value: 'overview', label: 'Visão Geral' },
  { value: 'conversations', label: 'Conversas' },
  { value: 'agents', label: 'Agentes' }
]

// ─── Data fetching ───────────────────────────────────────────────────────────
const { data: overview } = await useAsyncData('rep-overview', () =>
  api.get<{
    total_contacts: number
    open_conversations: number
    resolved_today: number
    online_agents: number
  }>('/api/reports/overview').catch(() => null)
)

const { data: convStats } = await useAsyncData('rep-conversations', () =>
  api.get<{ by_status: KV[]; by_inbox: KV[] }>('/api/reports/conversations').catch(() => null)
)

const { data: agents } = await useAsyncData('rep-agents', () =>
  api.get<AgentStat[]>('/api/reports/agents').catch(() => [] as AgentStat[])
)

const { data: timeseries, refresh: refreshTimeseries } = await useAsyncData(
  'rep-timeseries',
  () => api.get<TimeSeriesReport>('/api/reports/timeseries', {
    from: fromStr.value,
    to: toStr.value,
    group_by: groupBy.value
  }).catch(() => ({
    conversations: [],
    incoming_messages: [],
    outgoing_messages: [],
    resolutions: []
  } as TimeSeriesReport))
)

watch([fromStr, toStr, groupBy], () => refreshTimeseries())

// ─── Computed chart series ───────────────────────────────────────────────────
const convSeries = computed(() => [
  { data: timeseries.value?.conversations ?? [], label: 'Conversas', color: 'var(--ui-primary)' }
])

const msgSeries = computed(() => [
  { data: timeseries.value?.incoming_messages ?? [], label: 'Entrada', color: 'var(--ui-primary)' },
  { data: timeseries.value?.outgoing_messages ?? [], label: 'Saída', color: 'var(--ui-color-emerald-500)' }
])

const resSeries = computed(() => [
  { data: timeseries.value?.resolutions ?? [], label: 'Resoluções', color: 'var(--ui-color-emerald-500)' }
])

// ─── KPI totals for Conversations tab ───────────────────────────────────────
const totalConversations = computed(() =>
  (timeseries.value?.conversations ?? []).reduce((a, p) => a + p.value, 0)
)
const totalIncoming = computed(() =>
  (timeseries.value?.incoming_messages ?? []).reduce((a, p) => a + p.value, 0)
)
const totalOutgoing = computed(() =>
  (timeseries.value?.outgoing_messages ?? []).reduce((a, p) => a + p.value, 0)
)
const totalResolutions = computed(() =>
  (timeseries.value?.resolutions ?? []).reduce((a, p) => a + p.value, 0)
)

// ─── Agent table columns ─────────────────────────────────────────────────────
const agentColumns = [
  { accessorKey: 'agent', header: 'Agente' },
  { accessorKey: 'total', header: 'Total' },
  { accessorKey: 'resolved', header: 'Resolvidas' },
  { accessorKey: 'open', header: 'Abertas' }
]

// ─── Overview KPI cards ──────────────────────────────────────────────────────
const kpiCards = computed(() => [
  { label: 'Total Contatos', value: overview.value?.total_contacts ?? 0, icon: 'i-lucide-users', to: '/contacts' },
  { label: 'Conversas Abertas', value: overview.value?.open_conversations ?? 0, icon: 'i-lucide-message-square', to: '/conversations' },
  { label: 'Resolvidas Hoje', value: overview.value?.resolved_today ?? 0, icon: 'i-lucide-check-circle', to: '/conversations' },
  { label: 'Agentes Online', value: overview.value?.online_agents ?? 0, icon: 'i-lucide-circle-dot', to: '/settings/users' }
])
</script>

<template>
  <UDashboardPanel id="reports">
    <template #header>
      <UDashboardNavbar title="Relatórios">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
      </UDashboardNavbar>

      <UDashboardToolbar>
        <template #left>
          <HomeDateRangePicker v-model="range" />
          <HomePeriodSelect v-model="period" :range="range" />
        </template>
        <template #right>
          <UTabs v-model="tab" :items="tabItems" color="neutral" variant="link" />
        </template>
      </UDashboardToolbar>
    </template>

    <template #body>
      <!-- ── Visao Geral ─────────────────────────────────────────────── -->
      <template v-if="tab === 'overview'">
        <!-- KPI cards -->
        <UPageGrid class="lg:grid-cols-4 gap-4 sm:gap-6 lg:gap-px mb-6">
          <UPageCard
            v-for="card in kpiCards"
            :key="card.label"
            :icon="card.icon"
            :title="card.label"
            :to="card.to"
            variant="subtle"
            :ui="{
              container: 'gap-y-1.5',
              wrapper: 'items-start',
              leading: 'p-2.5 rounded-full bg-primary/10 ring ring-inset ring-primary/25 flex-col',
              title: 'font-normal text-muted text-xs uppercase'
            }"
            class="lg:rounded-none first:rounded-l-lg last:rounded-r-lg hover:z-1"
          >
            <span class="text-2xl font-semibold text-highlighted">
              {{ card.value.toLocaleString('pt-BR') }}
            </span>
          </UPageCard>
        </UPageGrid>

        <!-- Charts row -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
          <ReportsReportChart
            :series="convSeries"
            title="Conversas"
          />
          <ReportsReportChart
            :series="msgSeries"
            title="Mensagens"
          />
        </div>

        <!-- By status + by inbox -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <UCard>
            <template #header>
              <p class="text-sm font-medium text-highlighted">Por Status</p>
            </template>
            <div class="space-y-3">
              <div
                v-for="kv in (convStats?.by_status ?? [])"
                :key="kv.key"
                class="flex items-center justify-between"
              >
                <span class="text-sm capitalize text-default">{{ kv.key }}</span>
                <div class="flex items-center gap-2">
                  <div class="h-1.5 rounded-full bg-primary/20 w-24 overflow-hidden">
                    <div
                      class="h-full rounded-full bg-primary"
                      :style="{ width: `${Math.min(100, (kv.value / Math.max(...(convStats?.by_status ?? []).map(k => k.value), 1)) * 100)}%` }"
                    />
                  </div>
                  <span class="text-sm font-medium text-highlighted w-8 text-right">{{ kv.value }}</span>
                </div>
              </div>
              <p v-if="!convStats?.by_status?.length" class="text-sm text-muted text-center py-4">Sem dados</p>
            </div>
          </UCard>

          <UCard>
            <template #header>
              <p class="text-sm font-medium text-highlighted">Por Caixa de Entrada</p>
            </template>
            <div class="space-y-3">
              <div
                v-for="kv in (convStats?.by_inbox ?? [])"
                :key="kv.key"
                class="flex items-center justify-between"
              >
                <span class="text-sm text-default truncate max-w-[140px]">{{ kv.key }}</span>
                <div class="flex items-center gap-2">
                  <div class="h-1.5 rounded-full bg-primary/20 w-24 overflow-hidden">
                    <div
                      class="h-full rounded-full bg-primary"
                      :style="{ width: `${Math.min(100, (kv.value / Math.max(...(convStats?.by_inbox ?? []).map(k => k.value), 1)) * 100)}%` }"
                    />
                  </div>
                  <span class="text-sm font-medium text-highlighted w-8 text-right">{{ kv.value }}</span>
                </div>
              </div>
              <p v-if="!convStats?.by_inbox?.length" class="text-sm text-muted text-center py-4">Sem inboxes com conversas</p>
            </div>
          </UCard>
        </div>
      </template>

      <!-- ── Conversas ───────────────────────────────────────────────── -->
      <template v-else-if="tab === 'conversations'">
        <!-- 4 metric cards -->
        <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <ReportsReportMetricCard
            label="Conversas"
            :value="totalConversations"
            icon="i-lucide-message-square"
          />
          <ReportsReportMetricCard
            label="Msg. Recebidas"
            :value="totalIncoming"
            icon="i-lucide-arrow-down-left"
          />
          <ReportsReportMetricCard
            label="Msg. Enviadas"
            :value="totalOutgoing"
            icon="i-lucide-arrow-up-right"
          />
          <ReportsReportMetricCard
            label="Resoluções"
            :value="totalResolutions"
            icon="i-lucide-check-circle"
          />
        </div>

        <!-- Charts -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <ReportsReportChart
            :series="convSeries"
            title="Novas Conversas"
          />
          <ReportsReportChart
            :series="resSeries"
            title="Conversas Resolvidas"
          />
          <ReportsReportChart
            :series="msgSeries"
            title="Mensagens"
            class="lg:col-span-2"
            :height="200"
          />
        </div>
      </template>

      <!-- ── Agentes ─────────────────────────────────────────────────── -->
      <template v-else-if="tab === 'agents'">
        <UTable
          :data="agents ?? []"
          :columns="agentColumns"
          :ui="{
            base: 'table-fixed border-separate border-spacing-0',
            thead: '[&>tr]:bg-elevated/50 [&>tr]:after:content-none',
            tbody: '[&>tr]:last:[&>td]:border-b-0',
            th: 'py-2 first:rounded-l-lg last:rounded-r-lg border-y border-default first:border-l last:border-r',
            td: 'border-b border-default',
          }"
        />
        <p v-if="!agents?.length" class="text-center text-sm text-muted py-12">
          Nenhum agente encontrado.
        </p>
      </template>
    </template>
  </UDashboardPanel>
</template>
