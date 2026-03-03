<script setup lang="ts">
const route = useRoute()
const api = useApi()
const toast = useToast()
const id = route.params.id as string

interface Contact {
  id: string
  name: string
  email: string | null
  phone: string | null
  company: string | null
  avatar_url: string | null
  status: string
  score: number
  created_at: string
  updated_at: string
  custom_attributes: Record<string, string> | null
}

interface Conv {
  id: string
  status: string
  subject: string | null
  last_activity_at: string
}

const { data: contact, refresh } = await useAsyncData(`contact-${id}`, () =>
  api.get<Contact>(`/api/contacts/${id}`)
)

const { data: convs } = await useAsyncData(`convs-${id}`, () =>
  api.get<{ data: Conv[] }>(`/api/contacts/${id}/conversations`).catch(() => ({ data: [] }))
)

// ─── Relative time ────────────────────────────────────────────────────────────

function relativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 60) return `${mins} min atrás`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h atrás`
  const days = Math.floor(hrs / 24)
  if (days < 30) return `${days}d atrás`
  const months = Math.floor(days / 30)
  return `${months} ${months === 1 ? 'mês' : 'meses'} atrás`
}

// ─── Forms ────────────────────────────────────────────────────────────────────

const form = reactive({
  name:      '',
  company:   '',
  email:     '',
  phone:     '',
  city:      '',
  country:   '',
  biography: ''
})

const social = reactive({
  linkedin:  '',
  facebook:  '',
  instagram: '',
  twitter:   '',
  github:    ''
})

function syncForms(c: Contact | null) {
  if (!c) return
  form.name      = c.name
  form.company   = c.company ?? ''
  form.email     = c.email ?? ''
  form.phone     = c.phone ?? ''
  form.city      = c.custom_attributes?.city ?? ''
  form.country   = c.custom_attributes?.country ?? ''
  form.biography = c.custom_attributes?.biography ?? ''
  social.linkedin  = c.custom_attributes?.linkedin ?? ''
  social.facebook  = c.custom_attributes?.facebook ?? ''
  social.instagram = c.custom_attributes?.instagram ?? ''
  social.twitter   = c.custom_attributes?.twitter ?? ''
  social.github    = c.custom_attributes?.github ?? ''
}

watch(contact, (c) => syncForms(c ?? null), { immediate: true })

// ─── Save ─────────────────────────────────────────────────────────────────────

const saving = ref(false)

async function save() {
  saving.value = true
  try {
    await api.patch(`/api/contacts/${id}`, {
      name:    form.name,
      email:   form.email || null,
      phone:   form.phone || null,
      company: form.company || null,
      custom_attributes: {
        city:      form.city,
        country:   form.country,
        biography: form.biography,
        linkedin:  social.linkedin,
        facebook:  social.facebook,
        instagram: social.instagram,
        twitter:   social.twitter,
        github:    social.github
      }
    })
    toast.add({ title: 'Contato atualizado', color: 'success' })
    refresh()
  } catch {
    toast.add({ title: 'Erro ao atualizar', color: 'error' })
  } finally {
    saving.value = false
  }
}

async function saveCustomAttr(key: string, value: unknown) {
  const updated = { ...(contact.value?.custom_attributes ?? {}), [key]: value }
  await api.patch(`/api/contacts/${id}`, { custom_attributes: updated }).catch(() => {})
  refresh()
}

// ─── Delete ───────────────────────────────────────────────────────────────────

const confirmDelete = ref(false)
const deleting = ref(false)

async function remove() {
  deleting.value = true
  try {
    await api.del(`/api/contacts/${id}`)
    toast.add({ title: 'Contato removido', color: 'success' })
    navigateTo('/contacts')
  } catch {
    toast.add({ title: 'Erro ao remover', color: 'error' })
  } finally {
    deleting.value = false
  }
}

// ─── Right panel tabs ─────────────────────────────────────────────────────────

const rightTab = ref('history')

const rightTabs = [
  { label: 'Atributos', value: 'attributes' },
  { label: 'Histórico', value: 'history' },
  { label: 'Notas',     value: 'notes' }
]

// ─── Breadcrumb ───────────────────────────────────────────────────────────────

const breadcrumb = computed(() => [
  { label: 'Contatos', to: '/contacts' },
  { label: contact.value?.name ?? '...' }
])
</script>

<template>
  <!-- Left panel -->
  <UDashboardPanel id="contact-main">
    <template #header>
      <UDashboardNavbar>
        <template #leading>
          <UDashboardSidebarCollapse />
          <UBreadcrumb :items="breadcrumb" class="ml-2" />
        </template>
        <template #right>
          <UButton
            label="Bloquear contato"
            color="neutral"
            variant="outline"
            size="sm"
            icon="i-lucide-ban"
          />
          <UButton
            label="Enviar mensagem"
            size="sm"
            icon="i-lucide-send"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <!-- Profile header -->
      <div class="pb-5">
        <UAvatar
          :alt="contact?.name"
          :src="contact?.avatar_url ?? undefined"
          size="3xl"
          class="mb-3"
        />
        <h1 class="text-lg font-semibold text-highlighted">{{ contact?.name }}</h1>
        <p v-if="contact?.created_at" class="text-xs text-muted mt-0.5">
          Criado {{ relativeTime(contact.created_at) }}
          <template v-if="contact?.updated_at">
            &middot; Última atividade {{ relativeTime(contact.updated_at) }}
          </template>
        </p>
        <div class="mt-2">
          <UButton
            label="+ etiqueta"
            color="neutral"
            variant="outline"
            size="xs"
            icon="i-lucide-tag"
          />
        </div>
      </div>

      <USeparator />

      <!-- Alterar detalhes do contato -->
      <div class="py-5 space-y-3">
        <p class="text-xs font-semibold text-muted uppercase tracking-wider">
          Alterar detalhes do contato
        </p>

        <!-- Row 1: Nome + Empresa -->
        <div class="grid grid-cols-2 gap-2">
          <UInput v-model="form.name" placeholder="Digite o nome" />
          <UInput v-model="form.company" placeholder="Digite o nome da empresa" />
        </div>

        <!-- Row 2: Email + Telefone -->
        <div class="grid grid-cols-2 gap-2">
          <UInput v-model="form.email" type="email" placeholder="Digite o endereço de e-mail" />
          <UInput v-model="form.phone" placeholder="+55 Telefone" />
        </div>

        <!-- Row 3: Cidade + País -->
        <div class="grid grid-cols-2 gap-2">
          <UInput v-model="form.city" placeholder="Digite o nome da cidade" />
          <UInput v-model="form.country" placeholder="Selecione o país" />
        </div>

        <!-- Row 4: Biografia + Empresa (extra) -->
        <div class="grid grid-cols-2 gap-2">
          <UInput v-model="form.biography" placeholder="Digite uma biografia" />
          <div />
        </div>
      </div>

      <USeparator />

      <!-- Editar redes sociais -->
      <div class="py-5 space-y-3">
        <p class="text-xs font-semibold text-muted uppercase tracking-wider">
          Editar redes sociais
        </p>

        <div class="grid grid-cols-3 gap-2">
          <UInput v-model="social.linkedin" placeholder="Adicionar LinkedIn">
            <template #leading>
              <UIcon name="i-simple-icons-linkedin" class="size-4 text-muted" />
            </template>
          </UInput>
          <UInput v-model="social.facebook" placeholder="Adicionar Facebook">
            <template #leading>
              <UIcon name="i-simple-icons-facebook" class="size-4 text-muted" />
            </template>
          </UInput>
          <UInput v-model="social.instagram" placeholder="Adicionar Instagram">
            <template #leading>
              <UIcon name="i-simple-icons-instagram" class="size-4 text-muted" />
            </template>
          </UInput>
        </div>
        <div class="grid grid-cols-3 gap-2">
          <UInput v-model="social.twitter" placeholder="Adicionar Twitter">
            <template #leading>
              <UIcon name="i-simple-icons-x" class="size-4 text-muted" />
            </template>
          </UInput>
          <UInput v-model="social.github" placeholder="Adicionar Github">
            <template #leading>
              <UIcon name="i-simple-icons-github" class="size-4 text-muted" />
            </template>
          </UInput>
        </div>
      </div>

      <USeparator />

      <!-- Actions -->
      <div class="py-4 flex items-center gap-3">
        <UButton
          label="Atualizar contato"
          color="primary"
          :loading="saving"
          @click="save"
        />
        <UDropdownMenu
          :items="[[{
            label: 'Excluir contato',
            icon: 'i-lucide-trash-2',
            color: 'error' as const,
            onSelect: () => { confirmDelete = true }
          }]]"
          :content="{ align: 'start' }"
        >
          <UButton
            label="Excluir contato"
            color="neutral"
            variant="outline"
            trailing-icon="i-lucide-chevron-down"
          />
        </UDropdownMenu>
      </div>
    </template>
  </UDashboardPanel>

  <!-- Right panel -->
  <UDashboardPanel
    id="contact-right"
    :default-size="22"
    :min-size="18"
    :max-size="28"
    resizable
    class="hidden lg:flex border-l border-default"
  >
    <template #header>
      <UDashboardNavbar>
        <template #right>
          <UTabs
            v-model="rightTab"
            :items="rightTabs"
            variant="link"
            :content="false"
            size="sm"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <!-- Atributos -->
      <div v-if="rightTab === 'attributes'">
        <ContactsCustomAttributesPanel
          entity-type="contact"
          :entity-id="id"
          :values="(contact?.custom_attributes as Record<string, unknown>) ?? {}"
          @change="saveCustomAttr"
        />
      </div>

      <!-- Histórico de conversas -->
      <div v-else-if="rightTab === 'history'" class="space-y-2">
        <div v-if="!convs?.data?.length" class="text-center py-8 text-muted text-sm">
          Nenhuma conversa ainda.
        </div>
        <div
          v-for="c in convs?.data"
          :key="c.id"
          class="flex items-start gap-3 px-3 py-3 border-b border-default hover:bg-elevated/50 transition-colors cursor-pointer"
          @click="navigateTo('/conversations')"
        >
          <UAvatar :alt="contact?.name" :src="contact?.avatar_url ?? undefined" size="sm" class="shrink-0 mt-0.5" />
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between gap-1">
              <p class="text-sm font-medium truncate">{{ contact?.name }}</p>
              <span class="text-xs text-muted shrink-0">
                {{ relativeTime(c.last_activity_at) }}
              </span>
            </div>
            <p class="text-xs text-muted truncate mt-0.5">
              {{ c.subject ?? 'Sem assunto' }}
            </p>
          </div>
        </div>
      </div>

      <!-- Notas -->
      <div v-else-if="rightTab === 'notes'">
        <ContactsContactNotes :contact-id="id" />
      </div>
    </template>
  </UDashboardPanel>

  <!-- Delete modal -->
  <UModal v-model:open="confirmDelete">
    <template #content>
      <UCard>
        <template #header>
          <p class="font-semibold text-highlighted">Excluir contato</p>
        </template>
        <p class="text-sm text-muted">
          Tem certeza que deseja excluir <strong>{{ contact?.name }}</strong>?
          Esta ação não pode ser desfeita.
        </p>
        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton label="Cancelar" color="neutral" variant="ghost" @click="confirmDelete = false" />
            <UButton label="Excluir" color="error" :loading="deleting" @click="remove" />
          </div>
        </template>
      </UCard>
    </template>
  </UModal>
</template>
