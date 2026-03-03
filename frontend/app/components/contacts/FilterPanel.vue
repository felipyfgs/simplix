<script setup lang="ts">
export interface FilterCondition {
  field: string
  op: string
  value: string
}

const emit = defineEmits<{
  apply: [conditions: FilterCondition[]]
  clear: []
  close: []
}>()

const FIELD_OPTIONS = [
  { label: 'Nome',     value: 'name' },
  { label: 'Email',    value: 'email' },
  { label: 'Telefone', value: 'phone' },
  { label: 'Empresa',  value: 'company' },
  { label: 'Status',   value: 'status' }
]

const TEXT_OPS = [
  { label: 'Contém',       value: 'contains' },
  { label: 'Igual a',      value: 'equals' },
  { label: 'Começa com',   value: 'starts_with' },
  { label: 'Presente',     value: 'present' },
  { label: 'Não presente', value: 'not_present' }
]

const STATUS_OPS = [
  { label: 'Igual a', value: 'equals' }
]

const STATUS_VALUES = [
  { label: 'Novo',        value: 'novo' },
  { label: 'Qualificado', value: 'qualificado' },
  { label: 'Proposta',    value: 'proposta' },
  { label: 'Negociação',  value: 'negociacao' },
  { label: 'Fechado',     value: 'fechado' },
  { label: 'Perdido',     value: 'perdido' }
]

function makeRow(): FilterCondition {
  return { field: 'name', op: 'contains', value: '' }
}

const conditions = ref<FilterCondition[]>([makeRow()])

function addRow() {
  conditions.value.push(makeRow())
}

function removeRow(i: number) {
  if (conditions.value.length === 1) {
    conditions.value = [makeRow()]
  } else {
    conditions.value.splice(i, 1)
  }
}

function onFieldChange(i: number) {
  const c = conditions.value[i]
  if (!c) return
  if (c.field === 'status') {
    c.op = 'equals'
    c.value = 'novo'
  } else {
    c.op = 'contains'
    c.value = ''
  }
}

function opsFor(field: string) {
  return field === 'status' ? STATUS_OPS : TEXT_OPS
}

function hasInput(op: string) {
  return op !== 'present' && op !== 'not_present'
}

function clearAll() {
  conditions.value = [makeRow()]
  emit('clear')
}

function apply() {
  emit('apply', conditions.value.map(c => ({ ...c })))
}
</script>

<template>
  <div class="w-[540px] p-4 grid gap-4">
    <h3 class="text-sm font-semibold text-highlighted">Filtrar contatos</h3>

    <ul class="grid gap-2 list-none">
      <li v-for="(cond, i) in conditions" :key="i" class="flex items-center gap-2">
        <USelect
          v-model="cond.field"
          :items="FIELD_OPTIONS"
          value-key="value"
          class="w-32"
          @update:model-value="onFieldChange(i)"
        />

        <USelect
          v-model="cond.op"
          :items="opsFor(cond.field)"
          value-key="value"
          class="w-36"
        />

        <template v-if="hasInput(cond.op)">
          <USelect
            v-if="cond.field === 'status'"
            v-model="cond.value"
            :items="STATUS_VALUES"
            value-key="value"
            class="flex-1"
          />
          <UInput
            v-else
            v-model="cond.value"
            placeholder="Inserir valor"
            class="flex-1"
          />
        </template>
        <div v-else class="flex-1" />

        <UButton
          icon="i-lucide-trash-2"
          color="neutral"
          variant="ghost"
          size="sm"
          @click="removeRow(i)"
        />
      </li>
    </ul>

    <div class="flex items-center justify-between pt-1">
      <UButton
        label="Adicionar filtro"
        variant="link"
        icon="i-lucide-plus"
        color="primary"
        size="sm"
        @click="addRow"
      />
      <div class="flex gap-2">
        <UButton
          label="Limpar"
          color="neutral"
          variant="subtle"
          size="sm"
          @click="clearAll"
        />
        <UButton
          label="Aplicar"
          size="sm"
          @click="apply"
        />
      </div>
    </div>
  </div>
</template>
