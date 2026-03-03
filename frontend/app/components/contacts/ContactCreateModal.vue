<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

const emit = defineEmits<{ created: [] }>()

const api = useApi()
const toast = useToast()

const open = ref(false)
const loading = ref(false)

const schema = z.object({
  firstName: z.string().min(1, 'Obrigatório'),
  lastName:  z.string().optional(),
  email:     z.string().email('Email inválido').or(z.literal('')).optional(),
  phone:     z.string().optional(),
  city:      z.string().optional(),
  company:   z.string().optional(),
  bio:       z.string().optional(),
  linkedin:  z.string().optional(),
  facebook:  z.string().optional(),
  instagram: z.string().optional(),
  twitter:   z.string().optional(),
  github:    z.string().optional()
})

type Schema = z.output<typeof schema>

const state = reactive<Partial<Schema>>({
  firstName: '', lastName: '', email: '', phone: '',
  city: '', company: '', bio: '',
  linkedin: '', facebook: '', instagram: '', twitter: '', github: ''
})

const socialFields = [
  { key: 'linkedin'  as const, icon: 'i-simple-icons-linkedin',  placeholder: 'Adicionar LinkedIn'  },
  { key: 'facebook'  as const, icon: 'i-simple-icons-facebook',  placeholder: 'Adicionar Facebook'  },
  { key: 'instagram' as const, icon: 'i-simple-icons-instagram', placeholder: 'Adicionar Instagram' },
  { key: 'twitter'   as const, icon: 'i-simple-icons-x',         placeholder: 'Adicionar Twitter'   },
  { key: 'github'    as const, icon: 'i-simple-icons-github',    placeholder: 'Adicionar Github'    }
]

function resetState() {
  Object.assign(state, {
    firstName: '', lastName: '', email: '', phone: '',
    city: '', company: '', bio: '',
    linkedin: '', facebook: '', instagram: '', twitter: '', github: ''
  })
}

async function onSubmit(event: FormSubmitEvent<Schema>) {
  loading.value = true
  try {
    const d = event.data
    const name = [d.firstName, d.lastName].filter(Boolean).join(' ')
    const social = { linkedin: d.linkedin, facebook: d.facebook, instagram: d.instagram, twitter: d.twitter, github: d.github }
    const hasSocial = Object.values(social).some(v => v)
    const customAttributes: Record<string, unknown> = {}
    if (d.city)     customAttributes.city = d.city
    if (d.bio)      customAttributes.bio = d.bio
    if (hasSocial)  customAttributes.social_profiles = social

    await api.post('/api/contacts', {
      name,
      email:   d.email || null,
      phone:   d.phone || null,
      company: d.company || null,
      ...(Object.keys(customAttributes).length ? { custom_attributes: customAttributes } : {})
    })
    toast.add({ title: 'Contato salvo', color: 'success', icon: 'i-lucide-check-circle' })
    open.value = false
    resetState()
    emit('created')
  } catch {
    toast.add({ title: 'Erro ao criar contato', color: 'error', icon: 'i-lucide-x-circle' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    :ui="{ content: 'max-w-2xl', body: 'overflow-y-auto max-h-[80vh] sm:p-6', close: 'hidden' }"
  >
    <UButton label="Novo Contato" icon="i-lucide-user-plus" />

    <template #body>
      <UForm :schema="schema" :state="state" @submit="onSubmit">
        <!-- Title -->
        <p class="text-base font-semibold text-highlighted mb-5">Alterar detalhes do contato</p>

        <!-- Basic fields grid -->
        <div class="grid grid-cols-2 gap-3 mb-3">
          <!-- First name -->
          <UFormField name="firstName">
            <UInput
              v-model="state.firstName"
              placeholder="Digite o primeiro nome"
              class="w-full"
              :ui="{ base: 'bg-elevated border-0 ring-0 focus:ring-1' }"
            />
          </UFormField>

          <!-- Last name -->
          <UFormField name="lastName">
            <UInput
              v-model="state.lastName"
              placeholder="Digite o sobrenome"
              class="w-full"
              :ui="{ base: 'bg-elevated border-0 ring-0 focus:ring-1' }"
            />
          </UFormField>

          <!-- Email -->
          <UFormField name="email">
            <UInput
              v-model="state.email"
              type="email"
              placeholder="Digite o endereço de e-mail"
              class="w-full"
              :ui="{ base: 'bg-elevated border-0 ring-0 focus:ring-1' }"
            />
          </UFormField>

          <!-- Phone -->
          <UFormField name="phone">
            <UInput
              v-model="state.phone"
              placeholder="Digite o número de telefone"
              class="w-full"
              :ui="{ base: 'bg-elevated border-0 ring-0 focus:ring-1' }"
            />
          </UFormField>

          <!-- City -->
          <UFormField name="city">
            <UInput
              v-model="state.city"
              placeholder="Digite o nome da cidade"
              class="w-full"
              :ui="{ base: 'bg-elevated border-0 ring-0 focus:ring-1' }"
            />
          </UFormField>

          <!-- Company -->
          <UFormField name="company">
            <UInput
              v-model="state.company"
              placeholder="Digite o nome da empresa"
              class="w-full"
              :ui="{ base: 'bg-elevated border-0 ring-0 focus:ring-1' }"
            />
          </UFormField>

          <!-- Bio (full width) -->
          <UFormField name="bio" class="col-span-2">
            <UInput
              v-model="state.bio"
              placeholder="Digite uma biografia"
              class="w-full"
              :ui="{ base: 'bg-elevated border-0 ring-0 focus:ring-1' }"
            />
          </UFormField>
        </div>

        <!-- Social profiles -->
        <p class="text-sm font-semibold text-highlighted mb-3">Editar redes sociais</p>
        <div class="grid grid-cols-3 gap-2 mb-6">
          <div
            v-for="field in socialFields"
            :key="field.key"
            class="flex items-center gap-2 px-3 h-9 rounded-lg bg-elevated"
          >
            <UIcon :name="field.icon" class="size-4 text-muted shrink-0" />
            <input
              v-model="state[field.key]"
              :placeholder="field.placeholder"
              class="bg-transparent text-sm outline-none placeholder:text-muted text-default w-full min-w-0"
            />
          </div>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-between">
          <button
            type="button"
            class="text-sm font-medium text-primary hover:underline"
            @click="open = false"
          >
            Cancelar
          </button>
          <UButton
            label="Salvar contato"
            type="submit"
            :loading="loading"
          />
        </div>
      </UForm>
    </template>
  </UModal>
</template>
