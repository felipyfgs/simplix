<script setup lang="ts">
const actions = [{
  label: 'Ver Contatos',
  icon: 'i-lucide-users',
  to: '/contacts'
}, {
  label: 'Conversas',
  icon: 'i-lucide-message-square',
  to: '/conversations'
}, {
  label: 'Relatórios',
  icon: 'i-lucide-bar-chart-2',
  to: '/reports'
}, {
  label: 'Configurações',
  icon: 'i-lucide-settings',
  to: '/settings'
}]
</script>

<template>
  <div>
    <p class="text-xs font-semibold text-muted uppercase tracking-wider mb-3">Ações Rápidas</p>
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <UButton
        v-for="action in actions"
        :key="action.label"
        :to="action.to"
        :icon="action.icon"
        :label="action.label"
        variant="soft"
        block
      />
    </div>
  </div>
</template>