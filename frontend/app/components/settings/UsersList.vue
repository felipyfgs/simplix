<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

interface User {
  id: string
  name: string
  email: string
  role: string
}

defineProps<{ users: User[] }>()

const items = [{
  label: 'Editar usuário',
  icon: 'i-lucide-pencil'
}, {
  label: 'Remover usuário',
  icon: 'i-lucide-trash',
  color: 'error' as const
}] satisfies DropdownMenuItem[]
</script>

<template>
  <ul role="list" class="divide-y divide-default">
    <li
      v-for="user in users"
      :key="user.id"
      class="flex items-center justify-between gap-3 py-3 px-4 sm:px-6"
    >
      <div class="flex items-center gap-3 min-w-0">
        <UAvatar :alt="user.name" size="md" />
        <div class="text-sm min-w-0">
          <p class="text-highlighted font-medium truncate">{{ user.name }}</p>
          <p class="text-muted truncate">{{ user.email }}</p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <UBadge
          :color="user.role === 'admin' ? 'primary' : 'neutral'"
          :label="user.role"
          variant="subtle"
          size="sm"
        />
        <UDropdownMenu :items="items" :content="{ align: 'end' }">
          <UButton icon="i-lucide-ellipsis-vertical" color="neutral" variant="ghost" />
        </UDropdownMenu>
      </div>
    </li>
    <li v-if="users.length === 0" class="py-8 text-center text-sm text-muted">
      Nenhum usuário encontrado.
    </li>
  </ul>
</template>
