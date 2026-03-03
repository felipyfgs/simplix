<script setup lang="ts">
interface Webhook {
  id: string
  url: string
  subscriptions: string[]
  enabled: boolean
}

const api = useApi()
const toast = useToast()

const { data: webhooks, refresh } = await useAsyncData('settings-webhooks', () =>
  api.get<Webhook[]>('/api/webhooks').catch(() => [] as Webhook[])
)

const isOpen = ref(false)
const newURL = ref('')
const newSubs = ref<string[]>([])
const webhookEvents = [
  'contact.created',
  'contact.enriched',
  'conversation.created',
  'conversation.resolved',
  'message.created'
]

async function createWebhook() {
  await api.post('/api/webhooks', { url: newURL.value, subscriptions: newSubs.value })
  toast.add({ title: 'Webhook criado', icon: 'i-lucide-check', color: 'success' })
  isOpen.value = false
  newURL.value = ''
  newSubs.value = []
  refresh()
}

async function deleteWebhook(id: string) {
  await api.del(`/api/webhooks/${id}`)
  toast.add({ title: 'Webhook removido', color: 'success' })
  refresh()
}
</script>

<template>
  <div>
    <UPageCard
      title="Webhooks"
      description="Receba notificações em URLs externas quando eventos ocorrerem."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <UButton
        label="Novo Webhook"
        color="neutral"
        icon="i-lucide-plus"
        class="w-fit lg:ms-auto"
        @click="isOpen = true"
      />
    </UPageCard>

    <UPageCard variant="subtle" :ui="{ container: 'p-0 sm:p-0 gap-y-0', wrapper: 'items-stretch' }">
      <SettingsWebhooksList :webhooks="webhooks ?? []" @delete="deleteWebhook" />
    </UPageCard>

    <UModal v-model:open="isOpen" title="Novo Webhook">
      <template #body>
        <div class="space-y-4">
          <UFormField label="URL" required>
            <UInput v-model="newURL" placeholder="https://..." class="w-full" />
          </UFormField>
          <UFormField label="Eventos">
            <div class="space-y-2">
              <UCheckbox
                v-for="evt in webhookEvents"
                :key="evt"
                :label="evt"
                :model-value="newSubs.includes(evt)"
                @update:model-value="(v: boolean | 'indeterminate') => v === true ? newSubs.push(evt) : newSubs.splice(newSubs.indexOf(evt), 1)"
              />
            </div>
          </UFormField>
        </div>
      </template>
      <template #footer>
        <UButton color="neutral" variant="ghost" label="Cancelar" @click="isOpen = false" />
        <UButton label="Criar" @click="createWebhook" />
      </template>
    </UModal>
  </div>
</template>
