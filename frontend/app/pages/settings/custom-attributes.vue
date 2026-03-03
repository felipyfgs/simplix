<script setup lang="ts">
interface AttrDef {
  id: string
  entity_type: string
  attribute_key: string
  display_name: string
  attribute_type: string
  options: string[]
}

const api = useApi()
const toast = useToast()

const { data: attrs, refresh } = await useAsyncData('settings-custom-attrs', () =>
  api.get<AttrDef[]>('/api/custom-attributes').catch(() => [] as AttrDef[])
)

const isOpen = ref(false)
const newAttr = reactive({
  entity_type: 'contact',
  attribute_key: '',
  display_name: '',
  attribute_type: 'text',
  options: ''
})

const entityOptions = [
  { label: 'Contato', value: 'contact' },
  { label: 'Conversa', value: 'conversation' }
]
const attrTypeOptions = [
  { label: 'Texto', value: 'text' },
  { label: 'Número', value: 'number' },
  { label: 'Booleano', value: 'boolean' },
  { label: 'Data', value: 'date' },
  { label: 'Lista', value: 'list' }
]

async function createAttr() {
  const options = newAttr.attribute_type === 'list'
    ? newAttr.options.split(',').map(s => s.trim()).filter(Boolean)
    : []
  await api.post('/api/custom-attributes', {
    entity_type: newAttr.entity_type,
    attribute_key: newAttr.attribute_key,
    display_name: newAttr.display_name,
    attribute_type: newAttr.attribute_type,
    options
  })
  toast.add({ title: 'Campo criado', icon: 'i-lucide-check', color: 'success' })
  isOpen.value = false
  Object.assign(newAttr, { attribute_key: '', display_name: '', attribute_type: 'text', options: '' })
  refresh()
}

async function deleteAttr(id: string) {
  await api.del(`/api/custom-attributes/${id}`)
  toast.add({ title: 'Campo removido', color: 'success' })
  refresh()
}
</script>

<template>
  <div>
    <UPageCard
      title="Campos Customizados"
      description="Defina campos adicionais para contatos e conversas."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <UButton
        label="Novo Campo"
        color="neutral"
        icon="i-lucide-plus"
        class="w-fit lg:ms-auto"
        @click="isOpen = true"
      />
    </UPageCard>

    <UPageCard variant="subtle" :ui="{ container: 'p-0 sm:p-0 gap-y-0', wrapper: 'items-stretch' }">
      <SettingsCustomAttributesList :attrs="attrs ?? []" @delete="deleteAttr" />
    </UPageCard>

    <UModal v-model:open="isOpen" title="Novo Campo Customizado">
      <template #body>
        <div class="space-y-4">
          <UFormField label="Entidade" required>
            <USelect
              v-model="newAttr.entity_type"
              :items="entityOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>
          <UFormField label="Chave (snake_case)" required>
            <UInput v-model="newAttr.attribute_key" placeholder="ex: segmento_cliente" class="w-full" />
          </UFormField>
          <UFormField label="Nome de exibição" required>
            <UInput v-model="newAttr.display_name" placeholder="ex: Segmento" class="w-full" />
          </UFormField>
          <UFormField label="Tipo" required>
            <USelect
              v-model="newAttr.attribute_type"
              :items="attrTypeOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>
          <UFormField v-if="newAttr.attribute_type === 'list'" label="Opções (separadas por vírgula)">
            <UInput v-model="newAttr.options" placeholder="Opção 1, Opção 2" class="w-full" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <UButton color="neutral" variant="ghost" label="Cancelar" @click="isOpen = false" />
        <UButton label="Criar campo" @click="createAttr" />
      </template>
    </UModal>
  </div>
</template>
