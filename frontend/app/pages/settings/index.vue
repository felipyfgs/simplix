<script setup lang="ts">
const api = useApi()
const toast = useToast()

const { data: settings, refresh } = await useAsyncData('settings', () =>
  api.get<Record<string, string>>('/api/settings').catch(() => ({} as Record<string, string>))
)

const appName = ref('')
const saving = ref(false)

async function saveSettings() {
  if (!appName.value.trim()) return
  saving.value = true
  try {
    await api.patch('/api/settings', { app_name: appName.value })
    toast.add({ title: 'Configurações salvas', icon: 'i-lucide-check', color: 'success' })
    appName.value = ''
    refresh()
  }
  catch { toast.add({ title: 'Erro ao salvar', color: 'error' }) }
  finally { saving.value = false }
}
</script>

<template>
  <div>
    <UPageCard
      title="Geral"
      description="Configurações gerais da instância do Simplix CRM."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <UButton
        label="Salvar"
        color="neutral"
        :loading="saving"
        class="w-fit lg:ms-auto"
        @click="saveSettings"
      />
    </UPageCard>

    <UPageCard variant="subtle" :ui="{ container: 'divide-y divide-default' }">
      <UFormField
        label="Nome do app"
        description="Exibido na interface e em notificações."
        class="flex max-sm:flex-col justify-between items-start gap-4 not-last:pb-4"
      >
        <div class="flex flex-col gap-1 w-full sm:w-auto">
          <span class="text-sm font-medium text-highlighted">{{ settings?.app_name ?? 'Simplix CRM' }}</span>
          <UInput v-model="appName" placeholder="Novo nome..." autocomplete="off" />
        </div>
      </UFormField>
      <UFormField
        label="Backend URL"
        description="Endereço da API Go."
        class="flex max-sm:flex-col justify-between items-center gap-4 not-last:pb-4"
      >
        <span class="font-mono text-xs text-muted">{{ $config.public.apiBase }}</span>
      </UFormField>
    </UPageCard>
  </div>
</template>
