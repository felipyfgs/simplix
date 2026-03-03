<script setup lang="ts">
interface WaTemplate {
  name: string
  language: string
  status: string
  category: string
  components?: any[]
}

defineProps<{ templates: WaTemplate[] }>()

type BadgeColor = 'success' | 'warning' | 'error' | 'neutral' | 'primary' | 'info' | 'secondary'

const statusColor: Record<string, BadgeColor> = {
  APPROVED: 'success',
  PENDING: 'warning',
  REJECTED: 'error',
  PAUSED: 'neutral'
}
</script>

<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-default text-left">
          <th class="py-3 px-4 font-medium text-muted">Nome</th>
          <th class="py-3 px-4 font-medium text-muted">Idioma</th>
          <th class="py-3 px-4 font-medium text-muted">Categoria</th>
          <th class="py-3 px-4 font-medium text-muted">Status</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-default">
        <tr v-for="tpl in templates" :key="tpl.name + tpl.language" class="hover:bg-elevated/50">
          <td class="py-3 px-4 font-mono text-xs text-highlighted">{{ tpl.name }}</td>
          <td class="py-3 px-4 text-muted">{{ tpl.language }}</td>
          <td class="py-3 px-4 text-muted capitalize">{{ tpl.category?.toLowerCase() ?? '—' }}</td>
          <td class="py-3 px-4">
            <UBadge
              :label="tpl.status"
              :color="statusColor[tpl.status] ?? 'neutral'"
              variant="subtle"
              size="xs"
            />
          </td>
        </tr>
        <tr v-if="templates.length === 0">
          <td colspan="4" class="py-8 text-center text-muted">
            Nenhum template sincronizado. Clique em "Sincronizar" para buscar os templates aprovados.
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
