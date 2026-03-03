<script setup lang="ts">
import { formatDistanceToNow } from 'date-fns'
import { ptBR } from 'date-fns/locale'

interface Company {
  id: string
  name: string
  domain: string | null
  phone: string | null
  website: string | null
  industry: string | null
  description: string | null
  contacts_count: number
  updated_at: string
}

const props = defineProps<{ company: Company }>()

const initial = computed(() => props.company.name.charAt(0).toUpperCase())

const updatedAgo = computed(() => {
  if (!props.company.updated_at) return ''
  return formatDistanceToNow(new Date(props.company.updated_at), { addSuffix: true, locale: ptBR })
})
</script>

<template>
  <NuxtLink
    :to="`/companies/${company.id}`"
    class="flex items-center gap-4 px-4 py-3 border border-default rounded-lg bg-default hover:bg-elevated/50 transition-colors cursor-pointer"
  >
    <UAvatar :alt="company.name" size="md" class="shrink-0 text-sm font-semibold">
      {{ initial }}
    </UAvatar>

    <div class="flex-1 min-w-0">
      <div class="flex flex-wrap items-center gap-x-3 gap-y-0.5 min-w-0">
        <span class="text-sm font-semibold text-highlighted truncate">{{ company.name }}</span>
        <span v-if="company.industry" class="text-xs text-muted truncate">{{ company.industry }}</span>
      </div>

      <div class="flex flex-wrap items-center gap-x-3 gap-y-0.5 mt-0.5">
        <span v-if="company.description" class="text-xs text-muted truncate max-w-xs">
          {{ company.description }}
        </span>

        <span v-if="company.domain" class="inline-flex items-center gap-1 text-xs text-muted">
          <UIcon name="i-lucide-globe" class="size-3" />
          {{ company.domain }}
        </span>

        <template v-if="(company.description || company.domain) && company.contacts_count > 0">
          <span class="w-px h-3 bg-default-200 dark:bg-default-700" />
        </template>

        <span v-if="company.contacts_count > 0" class="inline-flex items-center gap-1 text-xs text-muted">
          <UIcon name="i-lucide-users" class="size-3" />
          {{ company.contacts_count }} {{ company.contacts_count === 1 ? 'contato' : 'contatos' }}
        </span>
      </div>
    </div>

    <span v-if="updatedAgo" class="text-xs text-muted shrink-0 hidden sm:block">
      {{ updatedAgo }}
    </span>

    <UIcon name="i-lucide-chevron-right" class="size-4 text-muted shrink-0" />
  </NuxtLink>
</template>
