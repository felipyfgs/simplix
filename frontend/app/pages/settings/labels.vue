<script setup lang="ts">
interface Label {
  id: string
  name: string
  color: string
  description: string | null
  created_at: string
}

const api = useApi()
const toast = useToast()

const { data: labels, refresh } = await useAsyncData('settings-labels', () =>
  api.get<Label[]>('/api/labels').catch(() => [] as Label[])
)

const isOpen = ref(false)
const editingLabel = ref<Label | null>(null)
const form = ref({ name: '', color: '#6B7280', description: '' })
const saving = ref(false)

const modalTitle = computed(() => editingLabel.value ? 'Editar Etiqueta' : 'Nova Etiqueta')

function openCreate() {
  editingLabel.value = null
  form.value = { name: '', color: '#6B7280', description: '' }
  isOpen.value = true
}

function openEdit(label: Label) {
  editingLabel.value = label
  form.value = { name: label.name, color: label.color, description: label.description ?? '' }
  isOpen.value = true
}

async function save() {
  if (!form.value.name) { toast.add({ title: 'Nome é obrigatório', color: 'error' }); return }
  saving.value = true
  try {
    const payload = { name: form.value.name, color: form.value.color, description: form.value.description || null }
    if (editingLabel.value) {
      await api.patch(`/api/labels/${editingLabel.value.id}`, payload)
      toast.add({ title: 'Etiqueta atualizada', color: 'success' })
    }
    else {
      await api.post('/api/labels', payload)
      toast.add({ title: 'Etiqueta criada', color: 'success' })
    }
    isOpen.value = false
    refresh()
  }
  catch (e: any) {
    toast.add({ title: 'Erro ao salvar', description: e?.data?.error ?? e?.message, color: 'error' })
  }
  finally { saving.value = false }
}

async function deleteLabel(id: string) {
  await api.del(`/api/labels/${id}`)
  toast.add({ title: 'Etiqueta removida', color: 'success' })
  refresh()
}

// Preset colors for the color picker
const presetColors = [
  '#EF4444', '#F97316', '#F59E0B', '#84CC16',
  '#10B981', '#06B6D4', '#3B82F6', '#8B5CF6',
  '#EC4899', '#6B7280', '#1E293B', '#FFFFFF'
]
</script>

<template>
  <div>
    <UPageCard
      title="Etiquetas"
      description="Crie etiquetas para organizar conversas e contatos."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <UButton
        label="Nova Etiqueta"
        color="neutral"
        icon="i-lucide-plus"
        class="w-fit lg:ms-auto"
        @click="openCreate"
      />
    </UPageCard>

    <UPageCard variant="subtle" :ui="{ container: 'p-0 sm:p-0 gap-y-0', wrapper: 'items-stretch' }">
      <div v-if="!labels?.length" class="flex flex-col items-center justify-center py-12 gap-2 text-center">
        <UIcon name="i-lucide-tag" class="size-8 text-muted" />
        <p class="text-sm text-muted">Nenhuma etiqueta criada ainda.</p>
      </div>

      <ul v-else class="divide-y divide-default">
        <li
          v-for="label in labels"
          :key="label.id"
          class="flex items-center gap-4 px-4 py-3"
        >
          <span
            class="size-3 rounded-full shrink-0"
            :style="{ backgroundColor: label.color }"
          />
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-highlighted">{{ label.name }}</p>
            <p v-if="label.description" class="text-xs text-muted truncate">{{ label.description }}</p>
          </div>
          <div class="flex items-center gap-1">
            <UButton
              icon="i-lucide-pencil"
              color="neutral"
              variant="ghost"
              size="xs"
              @click="openEdit(label)"
            />
            <UButton
              icon="i-lucide-trash"
              color="error"
              variant="ghost"
              size="xs"
              @click="deleteLabel(label.id)"
            />
          </div>
        </li>
      </ul>
    </UPageCard>

    <!-- Modal criar/editar -->
    <UModal v-model:open="isOpen" :title="modalTitle">
      <template #body>
        <div class="space-y-4">
          <UFormField label="Nome" required>
            <UInput v-model="form.name" placeholder="Ex: Urgente, VIP..." class="w-full" />
          </UFormField>

          <UFormField label="Cor">
            <div class="flex items-center gap-3">
              <span
                class="size-8 rounded-full border border-default shrink-0"
                :style="{ backgroundColor: form.color }"
              />
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="color in presetColors"
                  :key="color"
                  class="size-6 rounded-full border-2 transition-transform hover:scale-110"
                  :style="{ backgroundColor: color, borderColor: form.color === color ? 'var(--color-primary-500)' : 'transparent' }"
                  @click="form.color = color"
                />
              </div>
              <UInput v-model="form.color" placeholder="#6B7280" class="w-28" />
            </div>
          </UFormField>

          <UFormField label="Descrição">
            <UInput v-model="form.description" placeholder="Descrição opcional..." class="w-full" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <UButton color="neutral" variant="ghost" label="Cancelar" @click="isOpen = false" />
        <UButton :label="editingLabel ? 'Salvar' : 'Criar'" :loading="saving" @click="save" />
      </template>
    </UModal>
  </div>
</template>
