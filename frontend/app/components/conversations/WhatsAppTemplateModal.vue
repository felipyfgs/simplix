<script setup lang="ts">
interface WaTemplate {
  name: string
  language: string
  status: string
  category: string
  components?: Array<{
    type: string
    text?: string
    buttons?: any[]
  }>
}

const props = defineProps<{
  inboxId: string
  conversationId: string
}>()

const emit = defineEmits<{ close: [] }>()

const api = useApi()
const toast = useToast()

const { data: templates } = await useAsyncData(
  `tpl-${props.inboxId}`,
  () => api.get<WaTemplate[]>(`/api/inboxes/${props.inboxId}/templates`).catch(() => [] as WaTemplate[])
)

const approvedTemplates = computed(() =>
  (templates.value ?? []).filter(t => t.status === 'APPROVED')
)

const selected = ref<WaTemplate | null>(null)
const params = ref<Record<string, string>>({})
const sending = ref(false)

function extractParams(tpl: WaTemplate): string[] {
  const body = tpl.components?.find(c => c.type === 'BODY')
  if (!body?.text) return []
  const matches = body.text.match(/\{\{(\d+)\}\}/g) ?? []
  return [...new Set(matches.map(m => m.replace(/\{\{|\}\}/g, '')))]
}

const templateParams = computed(() => selected.value ? extractParams(selected.value) : [])

function select(tpl: WaTemplate) {
  selected.value = tpl
  params.value = {}
  templateParams.value.forEach(k => { params.value[k] = '' })
}

async function send() {
  if (!selected.value) return
  sending.value = true
  try {
    const components = []
    if (templateParams.value.length > 0) {
      components.push({
        type: 'body',
        parameters: templateParams.value.map(k => ({ type: 'text', text: params.value[k] ?? '' }))
      })
    }
    await api.post(`/api/conversations/${props.conversationId}/messages`, {
      content: `[template:${selected.value.name}]`,
      content_type: 'text',
      template: {
        name: selected.value.name,
        lang_code: selected.value.language,
        components
      }
    })
    toast.add({ title: 'Template enviado', icon: 'i-lucide-check', color: 'success' })
    emit('close')
  }
  catch { toast.add({ title: 'Erro ao enviar template', color: 'error' }) }
  finally { sending.value = false }
}

const bodyText = computed(() => {
  if (!selected.value) return ''
  const body = selected.value.components?.find(c => c.type === 'BODY')
  if (!body?.text) return ''
  let text = body.text
  templateParams.value.forEach(k => {
    text = text.replace(`{{${k}}}`, params.value[k] ? `*${params.value[k]}*` : `{{${k}}}`)
  })
  return text
})
</script>

<template>
  <UModal :open="true" title="Enviar Template WhatsApp" @update:open="(v) => !v && emit('close')">
    <template #body>
      <div v-if="!selected" class="space-y-2 max-h-80 overflow-y-auto">
        <p v-if="approvedTemplates.length === 0" class="text-sm text-muted py-4 text-center">
          Nenhum template aprovado. Sincronize os templates nas configurações da caixa de entrada.
        </p>
        <button
          v-for="tpl in approvedTemplates"
          :key="tpl.name + tpl.language"
          class="w-full text-left px-3 py-2.5 rounded-lg border border-default hover:border-primary hover:bg-primary/5 transition-colors"
          @click="select(tpl)"
        >
          <div class="flex items-center justify-between gap-2">
            <span class="font-mono text-sm font-medium text-highlighted">{{ tpl.name }}</span>
            <div class="flex items-center gap-1.5">
              <UBadge :label="tpl.language" variant="subtle" size="xs" />
              <UBadge :label="tpl.category?.toLowerCase()" variant="subtle" color="neutral" size="xs" />
            </div>
          </div>
          <p v-if="tpl.components?.find(c => c.type === 'BODY')?.text" class="text-xs text-muted mt-1 line-clamp-2">
            {{ tpl.components?.find(c => c.type === 'BODY')?.text }}
          </p>
        </button>
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center gap-2">
          <UButton icon="i-lucide-arrow-left" size="xs" color="neutral" variant="ghost" @click="selected = null" />
          <span class="font-mono text-sm font-semibold">{{ selected.name }}</span>
          <UBadge :label="selected.language" variant="subtle" size="xs" />
        </div>

        <!-- Preview -->
        <div class="bg-elevated border border-default rounded-lg p-3 text-sm whitespace-pre-wrap text-muted">
          {{ bodyText || '(sem corpo)' }}
        </div>

        <!-- Parameters -->
        <div v-if="templateParams.length > 0" class="space-y-3">
          <p class="text-xs font-medium text-muted uppercase tracking-wide">Parâmetros</p>
          <UFormField v-for="k in templateParams" :key="k" :label="`{{${k}}}`">
            <UInput v-model="params[k]" :placeholder="`Valor para {{${k}}}`" class="w-full" />
          </UFormField>
        </div>
      </div>
    </template>

    <template #footer>
      <UButton color="neutral" variant="ghost" label="Cancelar" @click="emit('close')" />
      <UButton
        v-if="selected"
        label="Enviar template"
        icon="i-lucide-send"
        :loading="sending"
        @click="send"
      />
    </template>
  </UModal>
</template>
