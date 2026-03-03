<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

definePageMeta({ layout: 'auth' })

const { signIn, isAuthenticated } = useAuth()

if (isAuthenticated.value) {
  await navigateTo('/')
}

const schema = z.object({
  email:    z.email('Email inválido'),
  password: z.string().min(1, 'Obrigatório')
})

type Schema = z.output<typeof schema>

const state = reactive<Schema>({ email: '', password: '' })
const loading = ref(false)
const errorMsg = ref('')

async function onSubmit(event: FormSubmitEvent<Schema>) {
  loading.value = true
  errorMsg.value = ''
  try {
    await signIn(event.data.email, event.data.password)
    await navigateTo('/')
  } catch {
    errorMsg.value = 'Email ou senha incorretos.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-6 p-4 w-full max-w-md">
    <!-- Logo -->
    <div class="flex items-center gap-2">
      <UIcon name="i-lucide-zap" class="size-7 text-primary" />
      <span class="text-xl font-bold text-highlighted">Simplix CRM</span>
    </div>

    <UPageCard class="w-full">
      <!-- Header -->
      <div class="flex flex-col items-center text-center mb-6">
        <div class="mb-3 p-3 rounded-full bg-primary/10">
          <UIcon name="i-lucide-lock" class="size-6 text-primary" />
        </div>
        <h1 class="text-xl font-semibold text-highlighted">Bem-vindo de volta</h1>
        <p class="text-sm text-muted mt-1">
          Não tem conta?
          <NuxtLink to="/register" class="text-primary font-medium hover:underline">Criar conta</NuxtLink>
        </p>
      </div>

      <UForm :schema="schema" :state="state" class="space-y-5" @submit="onSubmit">
        <UFormField label="Email" name="email" required>
          <UInput
            v-model="state.email"
            type="email"
            placeholder="admin@simplix.local"
            icon="i-lucide-mail"
            class="w-full"
            autocomplete="email"
          />
        </UFormField>

        <UFormField label="Senha" name="password" required>
          <UInput
            v-model="state.password"
            type="password"
            placeholder="••••••••"
            icon="i-lucide-lock"
            class="w-full"
            autocomplete="current-password"
          />
        </UFormField>

        <UAlert
          v-if="errorMsg"
          :description="errorMsg"
          color="error"
          variant="subtle"
          icon="i-lucide-alert-circle"
        />

        <UButton
          type="submit"
          label="Entrar"
          block
          :loading="loading"
        />
      </UForm>
    </UPageCard>
  </div>
</template>
