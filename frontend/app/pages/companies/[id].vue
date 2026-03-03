<script setup lang="ts">
interface Company {
  id: string
  name: string
  domain: string | null
  phone: string | null
  website: string | null
  industry: string | null
  description: string | null
  contacts_count: number
}

interface Contact {
  id: string
  name: string
  email: string | null
  phone: string | null
  avatar_url: string | null
  status: string
}

const route = useRoute()
const api = useApi()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: company, refresh } = await useAsyncData(
  () => `company-${id.value}`,
  () => api.get<Company>(`/api/companies/${id.value}`),
  { watch: [id] }
)

const { data: contacts } = await useAsyncData(
  () => `company-contacts-${id.value}`,
  () => api.get<Contact[]>(`/api/companies/${id.value}/contacts`).catch(() => [] as Contact[]),
  { watch: [id] }
)

const editing = ref(false)
const saving = ref(false)
const form = ref({ name: '', domain: '', phone: '', website: '', industry: '', description: '' })

watch(company, (c) => {
  if (c) form.value = {
    name: c.name,
    domain: c.domain ?? '',
    phone: c.phone ?? '',
    website: c.website ?? '',
    industry: c.industry ?? '',
    description: c.description ?? '',
  }
}, { immediate: true })

async function save() {
  saving.value = true
  try {
    await api.patch(`/api/companies/${id.value}`, {
      name: form.value.name,
      domain: form.value.domain || null,
      phone: form.value.phone || null,
      website: form.value.website || null,
      industry: form.value.industry || null,
      description: form.value.description || null,
    })
    toast.add({ title: 'Empresa atualizada', color: 'success' })
    editing.value = false
    refresh()
  }
  catch (e: any) {
    toast.add({ title: 'Erro ao salvar', description: e?.data?.error ?? e?.message, color: 'error' })
  }
  finally { saving.value = false }
}

const statusColor: Record<string, 'neutral' | 'info' | 'warning' | 'primary' | 'success' | 'error'> = {
  novo: 'neutral', qualificado: 'info', proposta: 'warning',
  negociacao: 'primary', fechado: 'success', perdido: 'error',
}

const infoFields = computed(() => {
  if (!company.value) return []
  return [
    { label: 'Domínio', value: company.value.domain, icon: 'i-lucide-globe' },
    { label: 'Telefone', value: company.value.phone, icon: 'i-lucide-phone' },
    { label: 'Website', value: company.value.website, icon: 'i-lucide-link' },
    { label: 'Setor', value: company.value.industry, icon: 'i-lucide-briefcase' },
  ]
})
</script>

<template>
  <UDashboardPanel v-if="company" id="company-detail">
    <template #header>
      <UDashboardNavbar :title="company.name">
        <template #leading>
          <UDashboardSidebarCollapse />
          <UButton
            icon="i-lucide-arrow-left"
            color="neutral"
            variant="ghost"
            to="/companies"
            aria-label="Voltar"
          />
        </template>
        <template #right>
          <template v-if="!editing">
            <UButton
              label="Editar"
              icon="i-lucide-pencil"
              color="neutral"
              variant="outline"
              size="sm"
              @click="editing = true"
            />
          </template>
          <template v-else>
            <UButton
              label="Cancelar"
              color="neutral"
              variant="ghost"
              size="sm"
              @click="editing = false"
            />
            <UButton
              label="Salvar"
              size="sm"
              :loading="saving"
              @click="save"
            />
          </template>
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <!-- Info section -->
      <div class="mb-8">
        <p class="text-xs font-semibold text-muted uppercase tracking-wider mb-4">Informações</p>

        <template v-if="!editing">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div v-for="field in infoFields" :key="field.label" class="flex items-start gap-3">
              <UIcon :name="field.icon" class="size-4 text-muted mt-0.5 shrink-0" />
              <div>
                <p class="text-xs text-muted">{{ field.label }}</p>
                <p class="text-sm text-highlighted">{{ field.value || '—' }}</p>
              </div>
            </div>
            <div v-if="company.description" class="sm:col-span-2 flex items-start gap-3">
              <UIcon name="i-lucide-align-left" class="size-4 text-muted mt-0.5 shrink-0" />
              <div>
                <p class="text-xs text-muted">Descrição</p>
                <p class="text-sm text-highlighted">{{ company.description }}</p>
              </div>
            </div>
          </div>
        </template>

        <template v-else>
          <div class="space-y-4 max-w-lg">
            <UFormField label="Nome" required>
              <UInput v-model="form.name" class="w-full" />
            </UFormField>
            <div class="grid grid-cols-2 gap-4">
              <UFormField label="Domínio">
                <UInput v-model="form.domain" placeholder="acme.com" class="w-full" />
              </UFormField>
              <UFormField label="Setor">
                <UInput v-model="form.industry" class="w-full" />
              </UFormField>
              <UFormField label="Telefone">
                <UInput v-model="form.phone" class="w-full" />
              </UFormField>
              <UFormField label="Website">
                <UInput v-model="form.website" placeholder="https://..." class="w-full" />
              </UFormField>
            </div>
            <UFormField label="Descrição">
              <UTextarea v-model="form.description" :rows="3" class="w-full" />
            </UFormField>
          </div>
        </template>
      </div>

      <!-- Contacts section -->
      <div>
        <p class="text-xs font-semibold text-muted uppercase tracking-wider mb-4">
          Contatos vinculados
          <span v-if="contacts?.length" class="ml-1 font-normal text-muted">({{ contacts.length }})</span>
        </p>

        <div v-if="!contacts?.length" class="flex flex-col items-center justify-center py-10 gap-2 border border-dashed border-default rounded-lg">
          <UIcon name="i-lucide-users" class="size-8 text-muted" />
          <p class="text-sm text-muted">Nenhum contato vinculado</p>
        </div>

        <div v-else class="flex flex-col gap-2">
          <NuxtLink
            v-for="c in contacts"
            :key="c.id"
            :to="`/contacts/${c.id}`"
            class="flex items-center gap-3 px-4 py-3 border border-default rounded-lg bg-default hover:bg-elevated/50 transition-colors"
          >
            <UAvatar
              :alt="c.name"
              :src="c.avatar_url ?? undefined"
              size="sm"
              class="shrink-0"
            />
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-highlighted truncate">{{ c.name }}</p>
              <p class="text-xs text-muted truncate">{{ c.email || c.phone || '—' }}</p>
            </div>
            <UBadge
              :color="statusColor[c.status] ?? 'neutral'"
              variant="subtle"
              size="xs"
              class="capitalize shrink-0"
            >
              {{ c.status }}
            </UBadge>
            <UIcon name="i-lucide-chevron-right" class="size-4 text-muted shrink-0" />
          </NuxtLink>
        </div>
      </div>
    </template>
  </UDashboardPanel>
</template>
