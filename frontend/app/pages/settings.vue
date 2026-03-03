<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const route = useRoute()

// Rotas principais — fora delas esconde toolbar e ajusta layout
const mainRoutes = ['/settings', '/settings/inboxes', '/settings/users', '/settings/webhooks', '/settings/custom-attributes', '/settings/labels', '/settings/profile']
const isMainRoute = computed(() => mainRoutes.includes(route.path))
const showToolbar = isMainRoute

// Mapa de segmentos para labels legíveis
const segmentLabels: Record<string, string> = {
  settings: 'Configurações',
  inboxes: 'Caixas de Entrada',
  new: 'Nova Caixa de Entrada',
  users: 'Usuários',
  webhooks: 'Webhooks',
  'custom-attributes': 'Campos Custom',
  labels: 'Etiquetas',
  profile: 'Perfil'
}

const breadcrumbItems = computed(() => {
  if (isMainRoute.value) return []
  const parts = route.path.split('/').filter(Boolean) // ['settings', 'inboxes', 'new']
  return parts.map((seg, i) => {
    const to = '/' + parts.slice(0, i + 1).join('/')
    const label = segmentLabels[seg] ?? seg
    const isLast = i === parts.length - 1
    return isLast ? { label } : { label, to }
  })
})

// Sub-páginas (wizard, detalhe): padding lateral mas sem py-12
const panelBodyUi = computed(() => isMainRoute.value ? { body: 'lg:py-12' } : { body: 'p-4 sm:p-6' })
const wrapperClass = computed(() =>
  isMainRoute.value
    ? 'flex flex-col gap-4 sm:gap-6 lg:gap-12 w-full lg:max-w-2xl mx-auto'
    : 'flex flex-col w-full max-w-5xl mx-auto'
)

const links = [[{
  label: 'Geral',
  icon: 'i-lucide-settings',
  to: '/settings',
  exact: true
}, {
  label: 'Perfil',
  icon: 'i-lucide-user',
  to: '/settings/profile'
}, {
  label: 'Caixas de Entrada',
  icon: 'i-lucide-inbox',
  to: '/settings/inboxes'
}, {
  label: 'Etiquetas',
  icon: 'i-lucide-tag',
  to: '/settings/labels'
}, {
  label: 'Usuários',
  icon: 'i-lucide-users',
  to: '/settings/users'
}, {
  label: 'Webhooks',
  icon: 'i-lucide-webhook',
  to: '/settings/webhooks'
}, {
  label: 'Campos Custom',
  icon: 'i-lucide-sliders-horizontal',
  to: '/settings/custom-attributes'
}]] satisfies NavigationMenuItem[][]
</script>

<template>
  <UDashboardPanel id="settings" :ui="panelBodyUi">
    <template #header>
      <UDashboardNavbar :title="isMainRoute ? 'Configurações' : ''">
        <template #leading>
          <UDashboardSidebarCollapse />
          <UBreadcrumb v-if="!isMainRoute" :items="breadcrumbItems" />
        </template>
      </UDashboardNavbar>

      <UDashboardToolbar v-if="showToolbar">
        <UNavigationMenu :items="links" highlight class="-mx-1 flex-1" />
      </UDashboardToolbar>
    </template>

    <template #body>
      <div :class="wrapperClass">
        <NuxtPage />
      </div>
    </template>
  </UDashboardPanel>
</template>
