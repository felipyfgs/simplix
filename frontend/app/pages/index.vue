<script setup lang="ts">
interface Overview {
  total_contacts: number
  open_conversations: number
  resolved_today: number
  online_agents: number
}

const api = useApi()

const { data: overview, refresh } = await useAsyncData('overview', () =>
  api.get<Overview>('/api/reports/overview').catch(() => ({
    total_contacts: 0,
    open_conversations: 0,
    resolved_today: 0,
    online_agents: 0
  }))
)

const stats = computed(() => [
  { label: 'Total Contatos', value: overview.value?.total_contacts ?? 0, icon: 'i-lucide-users', to: '/contacts' },
  { label: 'Conversas Abertas', value: overview.value?.open_conversations ?? 0, icon: 'i-lucide-message-square', to: '/conversations' },
  { label: 'Resolvidas Hoje', value: overview.value?.resolved_today ?? 0, icon: 'i-lucide-check-circle', to: '/conversations' },
  { label: 'Agentes Online', value: overview.value?.online_agents ?? 0, icon: 'i-lucide-circle-dot', to: '/settings/users' }
])

useIntervalFn(refresh, 30000)
</script>

<template>
  <UDashboardPanel id="home">
    <template #header>
      <UDashboardNavbar title="Dashboard">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <HomeDashboardStats :stats="stats" />
      <HomeDashboardActions />
    </template>
  </UDashboardPanel>
</template>
