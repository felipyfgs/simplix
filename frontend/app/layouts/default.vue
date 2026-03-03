<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const open = ref(false)
const { fetchMe, isAuthenticated } = useAuth()
const route = useRoute()

onMounted(() => fetchMe())

const links = computed(() => [[{
  label: 'Dashboard',
  icon: 'i-lucide-layout-dashboard',
  to: '/',
  active: route.path === '/',
  onSelect: () => { open.value = false }
}, {
  label: 'Contatos',
  icon: 'i-lucide-contact',
  to: '/contacts',
  active: route.path.startsWith('/contacts'),
  onSelect: () => { open.value = false }
}, {
  label: 'Empresas',
  icon: 'i-lucide-building-2',
  to: '/companies',
  active: route.path.startsWith('/companies'),
  onSelect: () => { open.value = false }
}, {
  label: 'Conversas',
  icon: 'i-lucide-message-square',
  to: '/conversations',
  active: route.path.startsWith('/conversations'),
  onSelect: () => { open.value = false }
}, {
  label: 'Relatórios',
  icon: 'i-lucide-bar-chart-2',
  to: '/reports',
  active: route.path.startsWith('/reports'),
  onSelect: () => { open.value = false }
}, {
  label: 'Configurações',
  icon: 'i-lucide-settings',
  to: '/settings',
  type: 'trigger' as const,
  defaultOpen: true,
  active: route.path.startsWith('/settings'),
  children: [{
    label: 'Geral',
    icon: 'i-lucide-settings',
    to: '/settings',
    active: route.path === '/settings',
    onSelect: () => { open.value = false }
  }, {
    label: 'Perfil',
    icon: 'i-lucide-user',
    to: '/settings/profile',
    active: route.path.startsWith('/settings/profile'),
    onSelect: () => { open.value = false }
  }, {
    label: 'Caixas de Entrada',
    icon: 'i-lucide-inbox',
    to: '/settings/inboxes',
    active: route.path.startsWith('/settings/inboxes'),
    onSelect: () => { open.value = false }
  }, {
    label: 'Etiquetas',
    icon: 'i-lucide-tag',
    to: '/settings/labels',
    active: route.path.startsWith('/settings/labels'),
    onSelect: () => { open.value = false }
  }, {
    label: 'Usuários',
    icon: 'i-lucide-users-round',
    to: '/settings/users',
    active: route.path.startsWith('/settings/users'),
    onSelect: () => { open.value = false }
  }, {
    label: 'Webhooks',
    icon: 'i-lucide-webhook',
    to: '/settings/webhooks',
    active: route.path.startsWith('/settings/webhooks'),
    onSelect: () => { open.value = false }
  }, {
    label: 'Campos Custom',
    icon: 'i-lucide-sliders-horizontal',
    to: '/settings/custom-attributes',
    active: route.path.startsWith('/settings/custom-attributes'),
    onSelect: () => { open.value = false }
  }]
}], [{
  label: 'API Docs',
  icon: 'i-lucide-code-2',
  to: 'http://localhost:8080/swagger/index.html',
  target: '_blank'
}]] satisfies NavigationMenuItem[][])

const groups = computed(() => [{
  id: 'links',
  label: 'Navegar para',
  items: links.value.flat()
}])
</script>

<template>
  <UDashboardGroup unit="rem">
    <UDashboardSidebar
      id="default"
      v-model:open="open"
      collapsible
      resizable
      class="bg-elevated/25"
      :ui="{ footer: 'lg:border-t lg:border-default' }"
    >
      <template #header="{ collapsed }">
        <TeamsMenu :collapsed="collapsed" />
      </template>

      <template #default="{ collapsed }">
        <UDashboardSearchButton :collapsed="collapsed" class="bg-transparent ring-default" />

        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[0]"
          orientation="vertical"
          tooltip
          popover
        />

        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[1]"
          orientation="vertical"
          tooltip
          class="mt-auto"
        />

      </template>

      <template #footer="{ collapsed }">
        <UserMenu :collapsed="collapsed" />
      </template>
    </UDashboardSidebar>

    <UDashboardSearch :groups="groups" />

    <slot />

    <NotificationsSlideover />
  </UDashboardGroup>
</template>
