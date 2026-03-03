<script setup lang="ts">
interface AttrDef {
  id: string
  entity_type: string
  attribute_key: string
  display_name: string
  attribute_type: string
  options: string[]
}

defineProps<{ attrs: AttrDef[] }>()
const emit = defineEmits<{ delete: [id: string] }>()

const entityColor = { contact: 'primary', conversation: 'secondary' } as const
const typeIcon = {
  text: 'i-lucide-text',
  number: 'i-lucide-hash',
  boolean: 'i-lucide-toggle-left',
  date: 'i-lucide-calendar',
  list: 'i-lucide-list'
} as const
</script>

<template>
  <ul role="list" class="divide-y divide-default">
    <li
      v-for="attr in attrs"
      :key="attr.id"
      class="flex items-center justify-between gap-3 py-3 px-4 sm:px-6"
    >
      <div class="flex items-center gap-3 min-w-0">
        <div class="p-2 rounded-lg bg-elevated">
          <UIcon
            :name="typeIcon[attr.attribute_type as keyof typeof typeIcon] ?? 'i-lucide-tag'"
            class="size-4 text-muted"
          />
        </div>
        <div class="min-w-0">
          <p class="text-sm font-medium text-highlighted truncate">{{ attr.display_name }}</p>
          <p class="text-xs text-muted font-mono">{{ attr.attribute_key }}</p>
        </div>
      </div>

      <div class="flex items-center gap-2 shrink-0">
        <UBadge
          :color="entityColor[attr.entity_type as keyof typeof entityColor] ?? 'neutral'"
          :label="attr.entity_type"
          variant="subtle"
          size="xs"
        />
        <UBadge :label="attr.attribute_type" variant="outline" size="xs" color="neutral" />
        <UButton
          icon="i-lucide-trash-2"
          size="xs"
          color="error"
          variant="ghost"
          @click="emit('delete', attr.id)"
        />
      </div>
    </li>
    <li v-if="attrs.length === 0" class="py-8 text-center text-sm text-muted">
      Nenhum campo customizado definido.
    </li>
  </ul>
</template>
