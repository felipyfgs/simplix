<script setup lang="ts">
const props = defineProps<{
  entityType: 'contact' | 'conversation'
  entityId: string
  values: Record<string, unknown>
}>()

const emit = defineEmits<{
  change: [key: string, value: unknown]
}>()

const api = useApi()

const { data: defs } = useAsyncData(`custom-attrs-${props.entityType}`, () =>
  api.get<AttrDef[]>('/api/custom-attributes', { entity_type: props.entityType }).catch(() => []),
  { server: false }
)

const localValues = reactive<Record<string, unknown>>({ ...props.values })

watch(() => props.values, (v) => {
  Object.assign(localValues, v)
}, { deep: true })

function update(key: string, value: unknown) {
  localValues[key] = value
  emit('change', key, value)
}

interface AttrDef {
  id: string
  attribute_key: string
  display_name: string
  attribute_type: string
  options: string[]
}
</script>

<template>
  <div v-if="(defs ?? []).length === 0" class="text-xs text-muted text-center py-2">
    Nenhum campo customizado configurado.
    <NuxtLink to="/settings" class="text-primary hover:underline">Configurar</NuxtLink>
  </div>

  <div v-else class="space-y-3">
    <div v-for="def in defs" :key="def.id">
      <label class="text-xs font-medium text-muted block mb-1">{{ def.display_name }}</label>

      <UInput
        v-if="def.attribute_type === 'text' || def.attribute_type === 'number'"
        :model-value="String(localValues[def.attribute_key] ?? '')"
        :type="def.attribute_type === 'number' ? 'number' : 'text'"
        size="sm"
        @update:model-value="update(def.attribute_key, $event)"
      />

      <UCheckbox
        v-else-if="def.attribute_type === 'boolean'"
        :model-value="Boolean(localValues[def.attribute_key])"
        :label="def.display_name"
        @update:model-value="update(def.attribute_key, $event)"
      />

      <UInput
        v-else-if="def.attribute_type === 'date'"
        :model-value="String(localValues[def.attribute_key] ?? '')"
        type="date"
        size="sm"
        @update:model-value="update(def.attribute_key, $event)"
      />

      <USelect
        v-else-if="def.attribute_type === 'list'"
        :model-value="String(localValues[def.attribute_key] ?? '')"
        :options="def.options"
        size="sm"
        @update:model-value="update(def.attribute_key, $event)"
      />
    </div>
  </div>
</template>
