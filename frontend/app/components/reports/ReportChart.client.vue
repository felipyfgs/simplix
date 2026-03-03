<script setup lang="ts">
import { VisXYContainer, VisLine, VisArea, VisAxis, VisCrosshair, VisTooltip } from '@unovis/vue'

interface Point {
  timestamp: string
  value: number
}

interface Series {
  data: Point[]
  label: string
  color: string
}

const props = defineProps<{
  series: Series[]
  title: string
  height?: number
}>()

const cardRef = useTemplateRef<HTMLElement | null>('cardRef')
const { width } = useElementSize(cardRef)

const colors = ['var(--ui-primary)', 'var(--ui-color-emerald-500)', 'var(--ui-color-amber-500)']

const total = computed(() =>
  props.series[0]?.data.reduce((acc, p) => acc + p.value, 0) ?? 0
)

const x = (_: Point, i: number) => i
const makeY = (seriesIndex: number) => (d: Point) => props.series[seriesIndex]?.data[props.series[0]?.data.indexOf(d) ?? -1]?.value ?? d.value

const xTicks = (i: number) => {
  const data = props.series[0]?.data
  if (!data || !data[i]) return ''
  if (i === 0 || i === data.length - 1) return data[i].timestamp
  if (data.length <= 14) return data[i].timestamp
  if (i % Math.ceil(data.length / 6) === 0) return data[i].timestamp
  return ''
}

const makeTemplate = (seriesIdx: number) => (d: Point) => {
  const s = props.series[seriesIdx]
  const val = s?.data.find(p => p.timestamp === d.timestamp)?.value ?? d.value
  return `${d.timestamp}: ${val}`
}
</script>

<template>
  <UCard ref="cardRef" :ui="{ root: 'overflow-visible', body: '!px-0 !pt-0 !pb-3' }">
    <template #header>
      <div>
        <p class="text-xs text-muted uppercase tracking-wide mb-1">{{ title }}</p>
        <p class="text-2xl font-semibold text-highlighted">{{ total.toLocaleString('pt-BR') }}</p>
      </div>
    </template>

    <div v-if="!series[0]?.data.length" class="flex items-center justify-center text-sm text-muted" :style="{ height: `${height ?? 240}px` }">
      Sem dados para o periodo
    </div>

    <VisXYContainer
      v-else
      :data="series[0].data"
      :padding="{ top: 20 }"
      :style="{ height: `${height ?? 240}px` }"
      :width="width"
    >
      <template v-for="(s, idx) in series" :key="idx">
        <VisLine
          :x="x"
          :y="(d: Point) => d.value"
          :color="colors[idx] ?? colors[0]"
        />
        <VisArea
          :x="x"
          :y="(d: Point) => d.value"
          :color="colors[idx] ?? colors[0]"
          :opacity="0.08"
        />
      </template>

      <VisAxis
        type="x"
        :x="x"
        :tick-format="xTicks"
      />

      <VisCrosshair
        :color="colors[0]"
        :template="makeTemplate(0)"
      />
      <VisTooltip />
    </VisXYContainer>
  </UCard>
</template>

<style scoped>
.unovis-xy-container {
  --vis-crosshair-line-stroke-color: var(--ui-primary);
  --vis-crosshair-circle-stroke-color: var(--ui-bg);
  --vis-axis-grid-color: var(--ui-border);
  --vis-axis-tick-color: var(--ui-border);
  --vis-axis-tick-label-color: var(--ui-text-dimmed);
  --vis-tooltip-background-color: var(--ui-bg);
  --vis-tooltip-border-color: var(--ui-border);
  --vis-tooltip-text-color: var(--ui-text-highlighted);
}
</style>
