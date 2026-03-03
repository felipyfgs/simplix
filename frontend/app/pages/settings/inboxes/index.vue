<script setup lang="ts">
interface Inbox {
  id: string
  name: string
  channel_type: string
  settings?: {
    phone_number: string
    phone_number_id: string
    business_account_id: string
    webhook_verify_token: string
  }
  created_at: string
}

const api = useApi()
const toast = useToast()

const { data: inboxes, refresh } = await useAsyncData('settings-inboxes', () =>
  api.get<Inbox[]>('/api/inboxes').catch(() => [] as Inbox[])
)

async function deleteInbox(id: string) {
  await api.del(`/api/inboxes/${id}`)
  toast.add({ title: 'Caixa removida', color: 'success' })
  refresh()
}
</script>

<template>
  <div>
    <UPageCard
      title="Caixas de Entrada"
      description="Gerencie os canais de comunicação conectados ao Simplix CRM."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <UButton
        label="Nova Caixa de Entrada"
        color="neutral"
        icon="i-lucide-plus"
        class="w-fit lg:ms-auto"
        to="/settings/inboxes/new"
      />
    </UPageCard>

    <UPageCard variant="subtle" :ui="{ container: 'p-0 sm:p-0 gap-y-0', wrapper: 'items-stretch' }">
      <SettingsInboxesList :inboxes="inboxes ?? []" @delete="deleteInbox" />
    </UPageCard>
  </div>
</template>
