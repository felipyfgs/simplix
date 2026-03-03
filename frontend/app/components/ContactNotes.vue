<script setup lang="ts">
const props = defineProps<{ contactId: string }>()
const api = useApi()
const toast = useToast()

const { data: notes, refresh } = useAsyncData(`notes-${props.contactId}`, () =>
  api.get<Note[]>(`/api/contacts/${props.contactId}/notes`).catch(() => []),
  { server: false }
)

const newNote = ref('')
const saving = ref(false)

async function addNote() {
  if (!newNote.value.trim()) return
  saving.value = true
  try {
    await api.post(`/api/contacts/${props.contactId}/notes`, { content: newNote.value })
    newNote.value = ''
    refresh()
    toast.add({ title: 'Nota adicionada', color: 'success' })
  }
  catch { toast.add({ title: 'Erro ao salvar nota', color: 'error' }) }
  finally { saving.value = false }
}

async function deleteNote(noteId: string) {
  await api.del(`/api/contacts/${props.contactId}/notes/${noteId}`)
  refresh()
}

interface Note {
  id: string
  content: string
  created_at: string
  author?: { name: string }
}
</script>

<template>
  <div class="space-y-3">
    <!-- Add note -->
    <div class="flex gap-2">
      <UTextarea
        v-model="newNote"
        placeholder="Adicionar nota..."
        :rows="2"
        class="flex-1 text-sm"
        @keydown.ctrl.enter="addNote"
      />
      <UButton
        icon="i-lucide-plus"
        :loading="saving"
        size="sm"
        variant="soft"
        @click="addNote"
      />
    </div>

    <!-- Notes list -->
    <div v-if="(notes ?? []).length === 0" class="text-center py-4 text-muted text-xs">
      Nenhuma nota ainda.
    </div>

    <div
      v-for="note in notes"
      :key="note.id"
      class="bg-warning-50 dark:bg-warning-900/20 border border-warning-200 dark:border-warning-800 rounded-lg p-3 text-sm relative group"
    >
      <p class="whitespace-pre-wrap">{{ note.content }}</p>
      <div class="flex items-center justify-between mt-2">
        <span class="text-xs text-muted">
          {{ note.author?.name ?? 'Sistema' }} •
          {{ new Date(note.created_at).toLocaleDateString('pt-BR') }}
        </span>
        <UButton
          icon="i-lucide-trash-2"
          size="xs"
          color="error"
          variant="ghost"
          class="opacity-0 group-hover:opacity-100"
          @click="deleteNote(note.id)"
        />
      </div>
    </div>
  </div>
</template>
