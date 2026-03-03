<script setup lang="ts">
interface User {
  id: string
  name: string
  email: string
  role: string
}

const api = useApi()

const { data: users } = await useAsyncData('settings-users', () =>
  api.get<User[]>('/api/users').catch(() => [] as User[])
)

const q = ref('')

const filteredUsers = computed(() =>
  (users.value ?? []).filter(u =>
    u.name.toLowerCase().includes(q.value.toLowerCase()) ||
    u.email.toLowerCase().includes(q.value.toLowerCase())
  )
)
</script>

<template>
  <div>
    <UPageCard
      title="Usuários"
      description="Gerencie os usuários com acesso ao Simplix CRM."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <UButton
        label="Convidar"
        color="neutral"
        icon="i-lucide-user-plus"
        class="w-fit lg:ms-auto"
      />
    </UPageCard>

    <UPageCard
      variant="subtle"
      :ui="{ container: 'p-0 sm:p-0 gap-y-0', wrapper: 'items-stretch', header: 'p-4 mb-0 border-b border-default' }"
    >
      <template #header>
        <UInput
          v-model="q"
          icon="i-lucide-search"
          placeholder="Buscar usuários..."
          autofocus
          class="w-full"
        />
      </template>

      <SettingsUsersList :users="filteredUsers" />
    </UPageCard>
  </div>
</template>
