<script setup lang="ts">
interface Conv {
  id: string
  contact_id: string
  status: string
  subject: string | null
}

interface Contact {
  id: string
  name: string
  email: string | null
  phone: string | null
  company: string | null
  avatar_url: string | null
  status: string
  custom_attributes: Record<string, unknown>
}

const props = defineProps<{ conversation: Conv }>()

const api = useApi()
const toast = useToast()

const { data: contact, refresh: refreshContact } = useAsyncData(
  () => `side-contact-${props.conversation.contact_id}`,
  () => api.get<Contact>(`/api/contacts/${props.conversation.contact_id}`).catch(() => null),
  { watch: [() => props.conversation.contact_id], server: false }
)

const activeTab = ref('info')
const tabs = [
  { value: 'info', label: 'Info' },
  { value: 'notes', label: 'Notas' },
  { value: 'custom', label: 'Custom' }
]

const statusColor = {
  novo: 'neutral', qualificado: 'info', proposta: 'warning',
  negociacao: 'primary', fechado: 'success', perdido: 'error'
} as const

async function saveCustomAttr(key: string, value: unknown) {
  if (!contact.value?.id) return
  const updated = { ...(contact.value.custom_attributes ?? {}), [key]: value }
  try {
    await api.patch(`/api/contacts/${contact.value.id}`, { custom_attributes: updated })
    refreshContact()
  }
  catch { toast.add({ title: 'Erro ao salvar atributo', color: 'error' }) }
}
</script>

<template>
  <UDashboardPanel id="conv-side" :default-size="22" :min-size="18" :max-size="30" resizable>
    <UDashboardNavbar :toggle="false">
      <template #leading>
        <div v-if="contact">
          <p class="font-semibold text-sm truncate">{{ contact.name }}</p>
          <p v-if="contact.company" class="text-xs text-muted">{{ contact.company }}</p>
        </div>
        <p v-else class="text-sm text-muted">Carregando...</p>
      </template>

      <template #right>
        <UBadge
          v-if="contact"
          :color="statusColor[contact.status as keyof typeof statusColor] ?? 'neutral'"
          :label="contact.status"
          variant="subtle"
          size="xs"
        />
      </template>
    </UDashboardNavbar>

    <div class="flex border-b border-default shrink-0">
      <button
        v-for="t in tabs"
        :key="t.value"
        class="flex-1 py-2 text-xs font-medium transition-colors"
        :class="activeTab === t.value
          ? 'text-primary border-b-2 border-primary'
          : 'text-muted hover:text-highlighted'"
        @click="activeTab = t.value"
      >
        {{ t.label }}
      </button>
    </div>

    <div class="flex-1 overflow-y-auto p-4">
      <!-- Info tab -->
      <dl v-if="activeTab === 'info' && contact" class="space-y-2 text-sm">
        <div v-if="contact.email" class="flex justify-between gap-2">
          <dt class="text-muted shrink-0">Email</dt>
          <dd class="truncate text-right">{{ contact.email }}</dd>
        </div>
        <div v-if="contact.phone" class="flex justify-between gap-2">
          <dt class="text-muted shrink-0">Telefone</dt>
          <dd>{{ contact.phone }}</dd>
        </div>
        <div v-if="contact.company" class="flex justify-between gap-2">
          <dt class="text-muted shrink-0">Empresa</dt>
          <dd class="truncate text-right">{{ contact.company }}</dd>
        </div>
        <div class="flex justify-between gap-2">
          <dt class="text-muted shrink-0">Status</dt>
          <dd class="capitalize">{{ contact.status }}</dd>
        </div>
        <div class="pt-2">
          <NuxtLink :to="`/contacts/${contact.id}`" class="text-xs text-primary hover:underline">
            Ver perfil completo →
          </NuxtLink>
        </div>
      </dl>

      <!-- Notes tab -->
      <div v-if="activeTab === 'notes'">
        <ContactsContactNotes :contact-id="conversation.contact_id" />
      </div>

      <!-- Custom tab -->
      <div v-if="activeTab === 'custom' && contact">
        <ContactsCustomAttributesPanel
          entity-type="contact"
          :entity-id="contact.id"
          :values="contact.custom_attributes ?? {}"
          @change="saveCustomAttr"
        />
      </div>
    </div>
  </UDashboardPanel>
</template>
