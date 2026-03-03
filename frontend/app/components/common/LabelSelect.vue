<script setup lang="ts">
interface Label {
  id: string
  name: string
  color: string
}

const props = defineProps<{
  modelValue: string[] // label IDs
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const api = useApi()

const { data: allLabels } = await useAsyncData('all-labels', () =>
  api.get<Label[]>('/api/labels').catch(() => [] as Label[])
)

const selected = computed(() => props.modelValue)

function toggleLabel(id: string) {
  const current = [...selected.value]
  const idx = current.indexOf(id)
  if (idx === -1) current.push(id)
  else current.splice(idx, 1)
  emit('update:modelValue', current)
}

const selectedLabels = computed(() =>
  (allLabels.value ?? []).filter(l => selected.value.includes(l.id))
)
</script>

<template>
  <div class="flex flex-wrap gap-1.5">
    <button
      v-for="label in allLabels"
      :key="label.id"
      :class="[
        'flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium border transition-all',
        selected.includes(label.id)
          ? 'border-transparent text-white'
          : 'border-default text-muted hover:border-default/80 bg-transparent'
      ]"
      :style="selected.includes(label.id) ? { backgroundColor: label.color } : {}"
      @click="toggleLabel(label.id)"
    >
      <span
        v-if="!selected.includes(label.id)"
        class="size-2 rounded-full"
        :style="{ backgroundColor: label.color }"
      />
      {{ label.name }}
    </button>

    <p v-if="!allLabels?.length" class="text-xs text-muted">Nenhuma etiqueta criada.</p>
  </div>
</template>
