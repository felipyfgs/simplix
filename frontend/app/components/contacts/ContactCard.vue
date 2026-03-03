<script setup lang="ts">
interface Contact {
  id: string
  name: string
  email: string | null
  phone: string | null
  company: string | null
  avatar_url: string | null
  status: string
  score: number
  custom_attributes: Record<string, string> | null
}

const props = defineProps<{ contact: Contact }>()
const emit = defineEmits<{ deleted: []; updated: [] }>()

const api = useApi()
const toast = useToast()

const expanded = ref(false)
const saving = ref(false)
const deleting = ref(false)
const confirmDelete = ref(false)

const form = reactive({
  name:      props.contact.name,
  email:     props.contact.email ?? '',
  phone:     props.contact.phone ?? '',
  company:   props.contact.company ?? '',
  city:      props.contact.custom_attributes?.city ?? '',
  country:   props.contact.custom_attributes?.country ?? '',
  biography: props.contact.custom_attributes?.biography ?? '',
  linkedin:  props.contact.custom_attributes?.linkedin ?? '',
  facebook:  props.contact.custom_attributes?.facebook ?? '',
  instagram: props.contact.custom_attributes?.instagram ?? '',
  twitter:   props.contact.custom_attributes?.twitter ?? '',
  github:    props.contact.custom_attributes?.github ?? ''
})

watch(() => props.contact, (c) => {
  form.name      = c.name
  form.email     = c.email ?? ''
  form.phone     = c.phone ?? ''
  form.company   = c.company ?? ''
  form.city      = c.custom_attributes?.city ?? ''
  form.country   = c.custom_attributes?.country ?? ''
  form.biography = c.custom_attributes?.biography ?? ''
  form.linkedin  = c.custom_attributes?.linkedin ?? ''
  form.facebook  = c.custom_attributes?.facebook ?? ''
  form.instagram = c.custom_attributes?.instagram ?? ''
  form.twitter   = c.custom_attributes?.twitter ?? ''
  form.github    = c.custom_attributes?.github ?? ''
})

async function save() {
  saving.value = true
  try {
    await api.patch(`/api/contacts/${props.contact.id}`, {
      name:    form.name,
      email:   form.email || null,
      phone:   form.phone || null,
      company: form.company || null,
      custom_attributes: {
        city:      form.city,
        country:   form.country,
        biography: form.biography,
        linkedin:  form.linkedin,
        facebook:  form.facebook,
        instagram: form.instagram,
        twitter:   form.twitter,
        github:    form.github
      }
    })
    toast.add({ title: 'Contato atualizado', color: 'success' })
    emit('updated')
  } catch {
    toast.add({ title: 'Erro ao atualizar contato', color: 'error' })
  } finally {
    saving.value = false
  }
}

async function remove() {
  deleting.value = true
  try {
    await api.del(`/api/contacts/${props.contact.id}`)
    toast.add({ title: 'Contato removido', color: 'success' })
    emit('deleted')
  } catch {
    toast.add({ title: 'Erro ao remover contato', color: 'error' })
  } finally {
    deleting.value = false
    confirmDelete.value = false
  }
}
</script>

<template>
  <div class="border border-default rounded-lg overflow-hidden bg-default">
    <!-- Row header (clicável) -->
    <div
      class="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-elevated/50 transition-colors"
      @click="expanded = !expanded"
    >
      <UAvatar
        :alt="contact.name"
        :src="contact.avatar_url ?? undefined"
        size="md"
        class="shrink-0"
      />

      <div class="flex-1 min-w-0">
        <p class="font-semibold text-sm text-highlighted truncate">{{ contact.name }}</p>
        <p class="text-xs text-muted truncate mt-0.5">
          <template v-if="contact.phone">{{ contact.phone }}</template>
          <template v-if="contact.phone && contact.email"> | </template>
          <template v-if="contact.email">{{ contact.email }}</template>
          <template v-if="!contact.phone && !contact.email && contact.company">{{ contact.company }}</template>
          <template v-if="!contact.phone && !contact.email && !contact.company">
            <span class="text-dimmed">Sem informações</span>
          </template>
        </p>
      </div>

      <NuxtLink
        :to="`/contacts/${contact.id}`"
        class="text-xs text-primary hover:underline shrink-0"
        @click.stop
      >
        Ver detalhes
      </NuxtLink>

      <UIcon
        :name="expanded ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
        class="size-4 text-muted shrink-0 transition-transform"
      />
    </div>

    <!-- Collapsible edit panel -->
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="opacity-0 max-h-0"
      enter-to-class="opacity-100 max-h-[800px]"
      leave-active-class="transition-all duration-150 ease-in"
      leave-from-class="opacity-100 max-h-[800px]"
      leave-to-class="opacity-0 max-h-0"
    >
      <div v-if="expanded" class="border-t border-default px-4 py-4 space-y-4">
        <p class="text-xs font-semibold text-muted uppercase tracking-wider">
          Alterar detalhes do contato
        </p>

        <!-- Row 1: Name + (display/emoji placeholder) -->
        <div class="grid grid-cols-2 gap-3">
          <UInput v-model="form.name" placeholder="Nome" />
          <UInput v-model="form.company" placeholder="Empresa" />
        </div>

        <!-- Row 2: Email + Phone -->
        <div class="grid grid-cols-2 gap-3">
          <UInput v-model="form.email" type="email" placeholder="Endereço de e-mail" />
          <UInput v-model="form.phone" placeholder="Telefone" />
        </div>

        <!-- Row 3: City + Country -->
        <div class="grid grid-cols-2 gap-3">
          <UInput v-model="form.city" placeholder="Cidade" />
          <UInput v-model="form.country" placeholder="País" />
        </div>

        <!-- Row 4: Biography -->
        <UTextarea v-model="form.biography" placeholder="Biografia" :rows="2" class="w-full" />

        <!-- Social networks -->
        <div>
          <p class="text-xs font-semibold text-muted uppercase tracking-wider mb-2">
            Editar redes sociais
          </p>
          <div class="grid grid-cols-3 gap-2">
            <UInput v-model="form.linkedin" placeholder="LinkedIn">
              <template #leading>
                <UIcon name="i-simple-icons-linkedin" class="size-4 text-muted" />
              </template>
            </UInput>
            <UInput v-model="form.facebook" placeholder="Facebook">
              <template #leading>
                <UIcon name="i-simple-icons-facebook" class="size-4 text-muted" />
              </template>
            </UInput>
            <UInput v-model="form.instagram" placeholder="Instagram">
              <template #leading>
                <UIcon name="i-simple-icons-instagram" class="size-4 text-muted" />
              </template>
            </UInput>
            <UInput v-model="form.twitter" placeholder="Twitter / X">
              <template #leading>
                <UIcon name="i-simple-icons-x" class="size-4 text-muted" />
              </template>
            </UInput>
            <UInput v-model="form.github" placeholder="Github">
              <template #leading>
                <UIcon name="i-simple-icons-github" class="size-4 text-muted" />
              </template>
            </UInput>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex items-center gap-3 pt-1">
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
              trailing-icon="i-lucide-chevron-down"
              color="neutral"
              variant="outline"
            />
          </UDropdownMenu>
        </div>
      </div>
    </Transition>

    <!-- Delete confirmation modal -->
    <UModal v-model:open="confirmDelete">
      <template #content>
        <UCard>
          <template #header>
            <p class="font-semibold text-highlighted">Excluir contato</p>
          </template>
          <p class="text-sm text-muted">
            Tem certeza que deseja excluir <strong>{{ contact.name }}</strong>? Esta ação não pode ser desfeita.
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
  </div>
</template>
