<script setup lang="ts">
interface StepperItem {
  title?: string
  description?: string
  icon?: string
}

const api = useApi()
const toast = useToast()
const config = useRuntimeConfig()

const currentStep = ref(0)

const steps = ref<StepperItem[]>([
  {
    title: 'Escolher canal',
    description: 'Escolha o provedor que deseja integrar.',
    icon: 'i-lucide-layout-grid'
  },
  {
    title: 'Criar caixa de entrada',
    description: 'Configure as credenciais e crie o canal.',
    icon: 'i-lucide-inbox'
  },
  {
    title: 'Pronto!',
    description: 'Tudo configurado para começar.',
    icon: 'i-lucide-check-circle'
  }
])

// Channels available
const channels = [
  {
    key: 'whatsapp',
    label: 'WhatsApp',
    description: 'Atenda seus clientes no WhatsApp via Cloud API',
    icon: 'i-lucide-message-circle',
    iconColor: 'text-success',
    available: true
  },
  {
    key: 'quepasa',
    label: 'WhatsApp (QuePasa)',
    description: 'WhatsApp via servidor QuePasa auto-hospedado',
    icon: 'i-lucide-message-circle-code',
    iconColor: 'text-success',
    available: true
  },
  {
    key: 'email',
    label: 'E-mail',
    description: 'Conectar com Gmail, Outlook ou outros provedores',
    icon: 'i-lucide-mail',
    iconColor: 'text-info',
    available: false
  },
  {
    key: 'api',
    label: 'API',
    description: 'Crie um canal personalizado usando nossa API',
    icon: 'i-lucide-code-2',
    iconColor: 'text-primary',
    available: false
  },
  {
    key: 'sms',
    label: 'SMS',
    description: 'Integre o canal SMS com Twilio ou outros',
    icon: 'i-lucide-smartphone',
    iconColor: 'text-warning',
    available: false
  },
  {
    key: 'telegram',
    label: 'Telegram',
    description: 'Configure o canal do Telegram usando o token do bot',
    icon: 'i-lucide-send',
    iconColor: 'text-primary',
    available: false
  },
  {
    key: 'instagram',
    label: 'Instagram',
    description: 'Conecte sua conta do Instagram',
    icon: 'i-lucide-instagram',
    iconColor: 'text-pink-500',
    available: false
  }
]

const selectedChannel = ref<string>('')

function selectChannel(key: string) {
  selectedChannel.value = key
  currentStep.value = 1
}

// WhatsApp Cloud API form
const waForm = ref({
  name: '',
  phone_number: '',
  phone_number_id: '',
  business_account_id: '',
  api_key: ''
})

// QuePasa form
const qpForm = ref({
  name: '',
  phone_number: '',
  bot_token: '',
  base_url: 'http://quepasa:31000'
})

const creating = ref(false)
const createdInbox = ref<{ id: string; name: string; channel_type?: string; settings?: { webhook_verify_token?: string } } | null>(null)

async function createInbox() {
  if (selectedChannel.value === 'quepasa') {
    await createQuePasaInbox()
  } else {
    await createWhatsAppInbox()
  }
}

async function createWhatsAppInbox() {
  if (!waForm.value.name || !waForm.value.phone_number || !waForm.value.phone_number_id || !waForm.value.business_account_id || !waForm.value.api_key) {
    toast.add({ title: 'Preencha todos os campos obrigatórios', color: 'error' })
    return
  }
  creating.value = true
  try {
    const result = await api.post<{ id: string; name: string; channel_type: string; settings?: { webhook_verify_token?: string } }>('/api/inboxes', waForm.value)
    createdInbox.value = result
    currentStep.value = 2
  }
  catch (e: any) {
    const msg = e?.data?.error ?? e?.message ?? 'Verifique as credenciais e tente novamente'
    toast.add({ title: 'Erro ao criar caixa de entrada', description: msg, color: 'error' })
  }
  finally { creating.value = false }
}

async function createQuePasaInbox() {
  if (!qpForm.value.name || !qpForm.value.phone_number || !qpForm.value.bot_token) {
    toast.add({ title: 'Preencha todos os campos obrigatórios', color: 'error' })
    return
  }
  creating.value = true
  try {
    const result = await api.post<{ id: string; name: string; channel_type: string }>('/api/inboxes', {
      name: qpForm.value.name,
      channel_type: 'quepasa',
      phone_number: qpForm.value.phone_number,
      bot_token: qpForm.value.bot_token,
      base_url: qpForm.value.base_url || 'http://quepasa:31000'
    })
    createdInbox.value = result
    currentStep.value = 2
  }
  catch (e: any) {
    const msg = e?.data?.error ?? e?.message ?? 'Verifique as credenciais e tente novamente'
    toast.add({ title: 'Erro ao criar caixa de entrada', description: msg, color: 'error' })
  }
  finally { creating.value = false }
}

const webhookUrl = computed(() => {
  if (!createdInbox.value) return ''
  const channel = createdInbox.value.channel_type ?? selectedChannel.value
  if (channel === 'quepasa') {
    return `${config.public.apiBase}/webhook/quepasa/${createdInbox.value.id}`
  }
  return `${config.public.apiBase}/webhook/whatsapp/${createdInbox.value.id}`
})

async function copyWebhookUrl() {
  await navigator.clipboard.writeText(webhookUrl.value)
  toast.add({ title: 'URL copiada', icon: 'i-lucide-check', color: 'success' })
}

async function copyVerifyToken() {
  const token = createdInbox.value?.settings?.webhook_verify_token ?? ''
  await navigator.clipboard.writeText(token)
  toast.add({ title: 'Token copiado', icon: 'i-lucide-check', color: 'success' })
}
</script>

<template>
  <div>
    <UPageCard
      title="Nova Caixa de Entrada"
      description="Conecte um canal de atendimento ao Simplix CRM."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <div class="flex items-center gap-2 lg:ms-auto">
        <UButton
          label="Voltar"
          color="neutral"
          variant="ghost"
          icon="i-lucide-arrow-left"
          to="/settings/inboxes"
        />
      </div>
    </UPageCard>

    <UPageCard variant="subtle" :ui="{ container: 'p-0 sm:p-0', wrapper: 'items-stretch' }">
    <div class="grid grid-cols-1 lg:grid-cols-8 overflow-hidden min-h-[60vh]">

      <!-- Esquerda: stepper vertical -->
      <div class="lg:col-span-2 border-b lg:border-b-0 lg:border-r border-default p-6 bg-elevated/25">
        <UStepper
          v-model="currentStep"
          :items="steps"
          orientation="vertical"
          color="neutral"
          disabled
          :ui="{
            separator: 'ms-[13px] my-0.5',
            wrapper: 'gap-1'
          }"
        />
      </div>

      <!-- Direita: conteúdo do passo -->
      <div class="lg:col-span-6 p-8 flex flex-col">

              <!-- Passo 1: Selecionar canal -->
              <template v-if="currentStep === 0">
                <h2 class="text-lg font-semibold text-highlighted mb-1">Escolha o Canal</h2>
                <p class="text-sm text-muted mb-6">
                  Escolha o provedor que deseja integrar com o Simplix CRM.
                </p>
                <div class="grid grid-cols-1 xs:grid-cols-2 sm:grid-cols-3 gap-4">
                  <button
                    v-for="ch in channels"
                    :key="ch.key"
                    :disabled="!ch.available"
                    :class="[
                      'relative flex flex-col items-center gap-3 p-5 rounded-xl border text-center transition-all',
                      ch.available
                        ? 'border-default hover:border-primary hover:bg-primary/5 cursor-pointer'
                        : 'border-dashed border-default opacity-50 cursor-not-allowed'
                    ]"
                    @click="selectChannel(ch.key)"
                  >
                    <div
                      :class="[
                        'size-12 rounded-xl flex items-center justify-center',
                        ch.available ? 'bg-elevated' : 'bg-elevated/50'
                      ]"
                    >
                      <UIcon :name="ch.icon" :class="['size-6', ch.iconColor]" />
                    </div>
                    <div>
                      <p class="font-medium text-sm text-highlighted">{{ ch.label }}</p>
                      <p class="text-xs text-muted mt-0.5 leading-relaxed">{{ ch.description }}</p>
                    </div>
                    <UBadge
                      v-if="!ch.available"
                      label="Em breve"
                      size="xs"
                      color="neutral"
                      variant="subtle"
                      class="absolute top-2 right-2"
                    />
                  </button>
                </div>
              </template>

              <!-- Passo 2: Formulário WhatsApp Cloud API -->
              <template v-else-if="currentStep === 1 && selectedChannel === 'whatsapp'">
                <h2 class="text-lg font-semibold text-highlighted mb-1">Configurar WhatsApp Cloud API</h2>
                <p class="text-sm text-muted mb-6">
                  Preencha os dados obtidos no
                  <a href="https://developers.facebook.com/apps/" target="_blank" class="text-primary underline">
                    painel de desenvolvedores Meta
                  </a>.
                </p>
                <div class="space-y-5 max-w-lg">
                  <UFormField label="Nome da caixa de entrada" required>
                    <UInput v-model="waForm.name" placeholder="Ex: WhatsApp Vendas" class="w-full" />
                  </UFormField>
                  <UFormField
                    label="Número de telefone"
                    description="Formato E.164 — ex: +5511999999999"
                    required
                  >
                    <UInput v-model="waForm.phone_number" placeholder="+5511999999999" class="w-full" />
                  </UFormField>
                  <UFormField
                    label="ID do número de telefone"
                    description="Phone Number ID obtido no painel de desenvolvedores Meta"
                    required
                  >
                    <UInput v-model="waForm.phone_number_id" placeholder="123456789012345" class="w-full" />
                  </UFormField>
                  <UFormField
                    label="ID da conta WhatsApp Business"
                    description="Business Account ID no painel Meta"
                    required
                  >
                    <UInput v-model="waForm.business_account_id" placeholder="987654321012345" class="w-full" />
                  </UFormField>
                  <UFormField
                    label="Chave da API"
                    description="Token de acesso permanente (System User Token)"
                    required
                  >
                    <UInput v-model="waForm.api_key" type="password" placeholder="EAAxxxxxx..." class="w-full" />
                  </UFormField>
                </div>

                <div class="flex items-center gap-3 mt-8">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    label="Voltar"
                    icon="i-lucide-arrow-left"
                    @click="currentStep = 0"
                  />
                  <UButton
                    label="Criar canal do WhatsApp"
                    icon="i-lucide-message-circle"
                    :loading="creating"
                    @click="createInbox"
                  />
                </div>
              </template>

              <!-- Passo 2: Formulário QuePasa -->
              <template v-else-if="currentStep === 1 && selectedChannel === 'quepasa'">
                <h2 class="text-lg font-semibold text-highlighted mb-1">Configurar WhatsApp via QuePasa</h2>
                <p class="text-sm text-muted mb-6">
                  Acesse a interface QuePasa em
                  <a href="http://localhost:31000" target="_blank" class="text-primary underline">localhost:31000</a>,
                  conecte um número via QR code e copie o token gerado.
                </p>
                <div class="space-y-5 max-w-lg">
                  <UFormField label="Nome da caixa de entrada" required>
                    <UInput v-model="qpForm.name" placeholder="Ex: WhatsApp QuePasa" class="w-full" />
                  </UFormField>
                  <UFormField
                    label="Número de telefone"
                    description="Formato E.164 — ex: +5511999999999"
                    required
                  >
                    <UInput v-model="qpForm.phone_number" placeholder="+5511999999999" class="w-full" />
                  </UFormField>
                  <UFormField
                    label="Token do bot"
                    description="Token exibido na interface QuePasa após conectar o número."
                    required
                  >
                    <UInput v-model="qpForm.bot_token" type="password" placeholder="token-do-bot..." class="w-full" />
                  </UFormField>
                  <UFormField
                    label="URL base do QuePasa"
                    description="URL interna do servidor QuePasa. Altere apenas se personalizado."
                  >
                    <UInput v-model="qpForm.base_url" placeholder="http://quepasa:31000" class="w-full" />
                  </UFormField>
                </div>

                <div class="flex items-center gap-3 mt-8">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    label="Voltar"
                    icon="i-lucide-arrow-left"
                    @click="currentStep = 0"
                  />
                  <UButton
                    label="Criar canal QuePasa"
                    icon="i-lucide-message-circle-code"
                    :loading="creating"
                    @click="createInbox"
                  />
                </div>
              </template>

              <!-- Passo 3: Concluído -->
              <template v-else>
                <div class="flex flex-col items-center text-center py-8 gap-6 max-w-lg mx-auto w-full">
                  <div class="size-16 rounded-full bg-success/10 flex items-center justify-center">
                    <UIcon name="i-lucide-check-circle" class="size-8 text-success" />
                  </div>
                  <div>
                    <h2 class="text-xl font-semibold text-highlighted">Caixa de entrada criada!</h2>
                    <p class="text-sm text-muted mt-1">
                      <strong class="text-highlighted">{{ createdInbox?.name }}</strong> está pronta.
                      Configure o webhook na Meta para começar a receber mensagens.
                    </p>
                  </div>

                  <div class="w-full text-left space-y-4">
                    <!-- Webhook URL -->
                    <div class="rounded-lg border border-default bg-elevated/50 p-4 space-y-2">
                      <p class="text-xs font-semibold text-muted uppercase tracking-wide">URL do Webhook</p>
                      <p v-if="selectedChannel === 'quepasa'" class="text-xs text-muted leading-relaxed">
                        Configure esta URL na interface QuePasa em
                        <a href="http://localhost:31000" target="_blank" class="text-primary underline">localhost:31000</a>
                        → Webhook. Use a URL interna do backend (ex: <code class="font-mono">http://backend:8080/webhook/quepasa/...</code>) se QuePasa e o backend estiverem no mesmo Docker network.
                      </p>
                      <p v-else class="text-xs text-muted leading-relaxed">
                        Cole esta URL no campo <strong>Callback URL</strong> do painel Meta Business Manager → WhatsApp → Configuração.
                      </p>
                      <div class="flex items-center gap-2 mt-2">
                        <code class="flex-1 text-xs font-mono bg-default border border-default rounded px-2 py-1.5 truncate">
                          {{ webhookUrl }}
                        </code>
                        <UButton
                          icon="i-lucide-copy"
                          size="xs"
                          color="neutral"
                          variant="ghost"
                          @click="copyWebhookUrl"
                        />
                      </div>
                    </div>

                    <!-- Verify token (WhatsApp only) -->
                    <div v-if="selectedChannel !== 'quepasa'" class="rounded-lg border border-default bg-elevated/50 p-4 space-y-2">
                      <p class="text-xs font-semibold text-muted uppercase tracking-wide">Token de Verificação</p>
                      <p class="text-xs text-muted leading-relaxed">
                        Cole este token no campo <strong>Verify Token</strong> ao configurar o webhook.
                      </p>
                      <div class="flex items-center gap-2 mt-2">
                        <code class="flex-1 text-xs font-mono bg-default border border-default rounded px-2 py-1.5 truncate">
                          {{ createdInbox?.settings?.webhook_verify_token ?? '—' }}
                        </code>
                        <UButton
                          icon="i-lucide-copy"
                          size="xs"
                          color="neutral"
                          variant="ghost"
                          @click="copyVerifyToken"
                        />
                      </div>
                    </div>
                  </div>

                  <UButton
                    label="Ver caixas de entrada"
                    icon="i-lucide-inbox"
                    color="neutral"
                    to="/settings/inboxes"
                    class="mt-2"
                  />
                </div>
              </template>

      </div>
    </div>
    </UPageCard>
  </div>
</template>
