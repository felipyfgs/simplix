<script setup lang="ts">
interface WhatsAppSettings {
  phone_number: string
  phone_number_id: string
  business_account_id: string
  api_key: string
  webhook_verify_token: string
  templates_last_synced?: string
}

interface QuePasaSettings {
  phone_number: string
  base_url: string
}

interface Inbox {
  id: string
  name: string
  channel_type: string
  settings?: WhatsAppSettings & QuePasaSettings
  created_at: string
}

const route = useRoute()
const api = useApi()
const toast = useToast()
const config = useRuntimeConfig()

const id = route.params.id as string

const { data: inbox, refresh } = await useAsyncData(`inbox-${id}`, () =>
  api.get<Inbox>(`/api/inboxes/${id}`)
)

const { data: templates, refresh: refreshTemplates } = await useAsyncData(`inbox-templates-${id}`, () =>
  api.get<any[]>(`/api/inboxes/${id}/templates`).catch(() => [] as any[])
)

const isQuePasa = computed(() => inbox.value?.channel_type === 'quepasa')
const isWhatsApp = computed(() => inbox.value?.channel_type === 'whatsapp')

const tabs = computed(() => {
  const t = [{ label: 'Configurações', value: 'settings', icon: 'i-lucide-settings' }]
  if (isWhatsApp.value) t.push({ label: 'Templates', value: 'templates', icon: 'i-lucide-layout-template' })
  return t
})
const activeTab = ref('settings')

const name = ref('')
// WhatsApp fields
const phoneNumber = ref('')
const phoneNumberId = ref('')
const businessAccountId = ref('')
const apiKey = ref('')
// QuePasa fields
const botToken = ref('')
const baseUrl = ref('')

const saving = ref(false)
const syncing = ref(false)

watch(inbox, (val) => {
  if (!val) return
  name.value = val.name
  phoneNumber.value = val.settings?.phone_number ?? ''
  phoneNumberId.value = val.settings?.phone_number_id ?? ''
  businessAccountId.value = val.settings?.business_account_id ?? ''
  baseUrl.value = val.settings?.base_url ?? 'http://quepasa:31000'
}, { immediate: true })

const quepasaWebhookUrl = computed(() =>
  inbox.value ? `${config.public.apiBase}/webhook/quepasa/${inbox.value.id}` : ''
)

async function copyQuePasaWebhookUrl() {
  await navigator.clipboard.writeText(quepasaWebhookUrl.value)
  toast.add({ title: 'URL copiada', icon: 'i-lucide-check', color: 'success' })
}

async function save() {
  saving.value = true
  try {
    const body: Record<string, string> = { name: name.value }
    if (isQuePasa.value) {
      if (phoneNumber.value) body.phone_number = phoneNumber.value
      if (botToken.value) body.bot_token = botToken.value
      if (baseUrl.value) body.base_url = baseUrl.value
    } else {
      if (phoneNumber.value) body.phone_number = phoneNumber.value
      if (phoneNumberId.value) body.phone_number_id = phoneNumberId.value
      if (businessAccountId.value) body.business_account_id = businessAccountId.value
      if (apiKey.value) body.api_key = apiKey.value
    }
    await api.patch(`/api/inboxes/${id}`, body)
    toast.add({ title: 'Configurações salvas', icon: 'i-lucide-check', color: 'success' })
    apiKey.value = ''
    botToken.value = ''
    refresh()
  }
  catch { toast.add({ title: 'Erro ao salvar', color: 'error' }) }
  finally { saving.value = false }
}

async function syncTemplates() {
  syncing.value = true
  try {
    const result = await api.post<{ synced: number }>(`/api/inboxes/${id}/templates/sync`)
    toast.add({ title: `${result.synced} templates sincronizados`, icon: 'i-lucide-check', color: 'success' })
    refreshTemplates()
    refresh()
  }
  catch { toast.add({ title: 'Erro ao sincronizar templates', color: 'error' }) }
  finally { syncing.value = false }
}
</script>

<template>
  <div>
    <UPageCard
      :title="inbox?.name ?? 'Caixa de Entrada'"
      :description="`Canal: ${inbox?.channel_type ?? '—'}`"
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <div class="flex items-center gap-2 lg:ms-auto">
        <UTabs
          v-if="tabs.length > 1"
          v-model="activeTab"
          :items="tabs"
          :content="false"
          size="sm"
          class="mr-2"
        />
        <UButton
          label="Voltar"
          color="neutral"
          variant="ghost"
          icon="i-lucide-arrow-left"
          to="/settings/inboxes"
        />
        <UButton
          v-if="activeTab === 'settings'"
          label="Salvar"
          color="neutral"
          :loading="saving"
          @click="save"
        />
        <UButton
          v-if="activeTab === 'templates'"
          label="Sincronizar"
          color="neutral"
          icon="i-lucide-refresh-cw"
          :loading="syncing"
          @click="syncTemplates"
        />
      </div>
    </UPageCard>

    <!-- Settings tab -->
    <UPageCard v-if="activeTab === 'settings'" variant="subtle" :ui="{ container: 'divide-y divide-default' }">
      <UFormField
        label="Nome da caixa de entrada"
        class="flex max-sm:flex-col justify-between items-start gap-4 not-last:pb-4"
      >
        <UInput v-model="name" placeholder="Nome..." class="w-full sm:w-72" />
      </UFormField>

      <!-- WhatsApp Cloud API settings -->
      <template v-if="isWhatsApp">
        <SettingsInboxWebhookInfo :inbox-id="id" class="not-last:pb-4" />

        <UFormField
          label="Token de verificação do webhook"
          description="Cole este token no campo Verify Token ao configurar o webhook na Meta."
          class="flex max-sm:flex-col justify-between items-start gap-4 not-last:pb-4"
        >
          <code class="text-xs font-mono bg-elevated border border-default rounded px-2 py-1.5">
            {{ inbox?.settings?.webhook_verify_token ?? '—' }}
          </code>
        </UFormField>

        <UFormField
          label="Número de telefone"
          class="flex max-sm:flex-col justify-between items-start gap-4 not-last:pb-4"
        >
          <UInput v-model="phoneNumber" placeholder="+5511999999999" class="w-full sm:w-72" />
        </UFormField>

        <UFormField
          label="Phone Number ID"
          class="flex max-sm:flex-col justify-between items-start gap-4 not-last:pb-4"
        >
          <UInput v-model="phoneNumberId" placeholder="123456789" class="w-full sm:w-72" />
        </UFormField>

        <UFormField
          label="Business Account ID"
          class="flex max-sm:flex-col justify-between items-start gap-4 not-last:pb-4"
        >
          <UInput v-model="businessAccountId" placeholder="987654321" class="w-full sm:w-72" />
        </UFormField>

        <UFormField
          label="Chave da API"
          description="Deixe em branco para manter a chave atual."
          class="flex max-sm:flex-col justify-between items-start gap-4"
        >
          <UInput v-model="apiKey" type="password" placeholder="Nova chave (opcional)..." class="w-full sm:w-72" />
        </UFormField>
      </template>

      <!-- QuePasa settings -->
      <template v-else-if="isQuePasa">
        <UFormField
          label="URL do webhook"
          description="Configure esta URL no painel QuePasa (localhost:31000) → Webhook. Use a URL interna do backend no Docker network."
          class="flex max-sm:flex-col justify-between items-start gap-4 not-last:pb-4"
        >
          <div class="flex items-center gap-2 w-full sm:w-auto">
            <code class="text-xs font-mono bg-elevated border border-default rounded px-2 py-1.5 truncate max-w-xs">
              {{ quepasaWebhookUrl }}
            </code>
            <UButton icon="i-lucide-copy" size="xs" color="neutral" variant="ghost" @click="copyQuePasaWebhookUrl" />
          </div>
        </UFormField>

        <UFormField
          label="Número de telefone"
          class="flex max-sm:flex-col justify-between items-start gap-4 not-last:pb-4"
        >
          <UInput v-model="phoneNumber" placeholder="+5511999999999" class="w-full sm:w-72" />
        </UFormField>

        <UFormField
          label="Token do bot"
          description="Deixe em branco para manter o token atual."
          class="flex max-sm:flex-col justify-between items-start gap-4 not-last:pb-4"
        >
          <UInput v-model="botToken" type="password" placeholder="Novo token (opcional)..." class="w-full sm:w-72" />
        </UFormField>

        <UFormField
          label="URL base do QuePasa"
          class="flex max-sm:flex-col justify-between items-start gap-4"
        >
          <UInput v-model="baseUrl" placeholder="http://quepasa:31000" class="w-full sm:w-72" />
        </UFormField>
      </template>
    </UPageCard>

    <!-- Templates tab -->
    <UPageCard
      v-if="activeTab === 'templates'"
      variant="subtle"
      :ui="{ container: 'p-0 sm:p-0', wrapper: 'items-stretch' }"
    >
      <SettingsInboxTemplatesList :templates="(templates as any[]) ?? []" />
    </UPageCard>
  </div>
</template>
