<script setup lang="ts">
interface AuthUser {
  id: string
  name: string
  email: string
  role: string
  availability: string
  avatar_url: string | null
}

const api = useApi()
const toast = useToast()
const { user: rawUser, fetchMe } = useAuth()
const user = computed(() => rawUser.value as AuthUser | null)

const form = ref({
  name: (rawUser.value as AuthUser | null)?.name ?? '',
  availability: (rawUser.value as AuthUser | null)?.availability ?? 'offline',
  avatar_url: (rawUser.value as AuthUser | null)?.avatar_url ?? ''
})

watch(rawUser, (u) => {
  const typed = u as AuthUser | null
  if (typed) {
    form.value.name = typed.name
    form.value.availability = typed.availability
    form.value.avatar_url = typed.avatar_url ?? ''
  }
}, { immediate: true })

const saving = ref(false)

const availabilityOptions = [
  { label: 'Online', value: 'online' },
  { label: 'Ocupado', value: 'busy' },
  { label: 'Offline', value: 'offline' }
]

async function save() {
  if (!form.value.name) { toast.add({ title: 'Nome é obrigatório', color: 'error' }); return }
  saving.value = true
  try {
    await api.patch('/api/auth/profile', {
      name: form.value.name,
      availability: form.value.availability,
      avatar_url: form.value.avatar_url || null
    })
    await fetchMe()
    toast.add({ title: 'Perfil atualizado', icon: 'i-lucide-check', color: 'success' })
  }
  catch (e: any) {
    toast.add({ title: 'Erro ao salvar', description: e?.data?.error ?? e?.message, color: 'error' })
  }
  finally { saving.value = false }
}
</script>

<template>
  <div>
    <UPageCard
      title="Perfil"
      description="Suas informações pessoais exibidas para a equipe."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <UButton
        label="Salvar"
        color="neutral"
        :loading="saving"
        class="w-fit lg:ms-auto"
        @click="save"
      />
    </UPageCard>

    <UPageCard variant="subtle">
      <UFormField
        label="Nome"
        description="Exibido na interface e nos atendimentos."
        required
        class="flex max-sm:flex-col justify-between items-start gap-4"
      >
        <UInput v-model="form.name" placeholder="Seu nome" autocomplete="off" />
      </UFormField>

      <USeparator />

      <UFormField
        label="E-mail"
        description="Usado para login. Não pode ser alterado aqui."
        class="flex max-sm:flex-col justify-between items-start gap-4"
      >
        <UInput :model-value="user?.email ?? ''" disabled class="opacity-60" />
      </UFormField>

      <USeparator />

      <UFormField
        label="Disponibilidade"
        description="Define seu status visível para a equipe."
        class="flex max-sm:flex-col justify-between items-start gap-4"
      >
        <USelect
          v-model="form.availability"
          :options="availabilityOptions"
          value-key="value"
          label-key="label"
          class="w-48"
        />
      </UFormField>

      <USeparator />

      <UFormField
        label="URL do Avatar"
        description="Link direto para uma imagem de perfil (JPG, PNG)."
        class="flex max-sm:flex-col justify-between items-start gap-4"
      >
        <div class="flex items-center gap-3">
          <UAvatar
            :src="form.avatar_url || undefined"
            :fallback="form.name ? form.name.charAt(0).toUpperCase() : 'U'"
            size="lg"
          />
          <UInput v-model="form.avatar_url" placeholder="https://..." class="w-64" />
        </div>
      </UFormField>
    </UPageCard>
  </div>
</template>
