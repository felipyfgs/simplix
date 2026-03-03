<script setup lang="ts">
interface Stat {
  label: string
  value: number
  icon: string
  to: string
}

defineProps<{ stats: Stat[] }>()
</script>

<template>
  <UPageGrid class="lg:grid-cols-4 gap-4 sm:gap-6 lg:gap-px">
    <UPageCard
      v-for="stat in stats"
      :key="stat.label"
      :icon="stat.icon"
      :title="stat.label"
      :to="stat.to"
      variant="subtle"
      :ui="{
        container: 'gap-y-1.5',
        wrapper: 'items-start',
        leading: 'p-2.5 rounded-full bg-primary/10 ring ring-inset ring-primary/25 flex-col',
        title: 'font-normal text-muted text-xs uppercase'
      }"
      class="lg:rounded-none first:rounded-l-lg last:rounded-r-lg hover:z-1"
    >
      <span class="text-2xl font-semibold text-highlighted">
        {{ stat.value.toLocaleString('pt-BR') }}
      </span>
    </UPageCard>
  </UPageGrid>
</template>
