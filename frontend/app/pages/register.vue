<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

definePageMeta({ layout: 'auth' })

const { isAuthenticated, signUp } = useAuth()
const toast = useToast()

if (isAuthenticated.value) {
  await navigateTo('/')
}

const schema = z.object({
  name:            z.string().min(2, 'Mínimo 2 caracteres'),
  email:           z.email('Email inválido'),
  password:        z.string().min(8, 'Mínimo 8 caracteres'),
  confirmPassword: z.string().min(1, 'Obrigatório')
}).refine(d => d.password === d.confirmPassword, {
  message: 'As senhas não coincidem',
  path: ['confirmPassword']
})

type Schema = z.output<typeof schema>

const state = reactive<Schema>({ name: '', email: '', password: '', confirmPassword: '' })
const loading = ref(false)
const errorMsg = ref('')

async function onSubmit(event: FormSubmitEvent<Schema>) {
  loading.value = true
  errorMsg.value = ''
  try {
    await signUp(event.data.name, event.data.email, event.data.password)
    toast.add({ title: 'Conta criada com sucesso!', color: 'success' })
    await navigateTo('/')
  } catch (e: unknown) {
    const err = e as { data?: { error?: string } }
    errorMsg.value = err.data?.error ?? 'Erro ao criar conta. Tente novamente.'
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
          <UIcon name="i-lucide-user-plus" class="size-6 text-primary" />
        </div>
        <h1 class="text-xl font-semibold text-highlighted">Criar conta</h1>
        <p class="text-sm text-muted mt-1">
          Já tem uma conta?
          <NuxtLink to="/login" class="text-primary font-medium hover:underline">Entrar</NuxtLink>
        </p>
      </div>

      <UForm :schema="schema" :state="state" class="space-y-4" @submit="onSubmit">
        <UFormField label="Nome" name="name" required>
          <UInput
            v-model="state.name"
            placeholder="Seu nome completo"
            icon="i-lucide-user"
            class="w-full"
            autocomplete="name"
          />
        </UFormField>

        <UFormField label="Email" name="email" required>
          <UInput
            v-model="state.email"
            type="email"
            placeholder="seu@email.com"
            icon="i-lucide-mail"
            class="w-full"
            autocomplete="email"
          />
        </UFormField>

        <UFormField label="Senha" name="password" required>
          <UInput
            v-model="state.password"
            type="password"
            placeholder="Mínimo 8 caracteres"
            icon="i-lucide-lock"
            class="w-full"
            autocomplete="new-password"
          />
        </UFormField>

        <UFormField label="Confirmar senha" name="confirmPassword" required>
          <UInput
            v-model="state.confirmPassword"
            type="password"
            placeholder="Repita a senha"
            icon="i-lucide-lock"
            class="w-full"
            autocomplete="new-password"
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
          label="Criar conta"
          block
          :loading="loading"
        />
      </UForm>
    </UPageCard>
  </div>
</template>
