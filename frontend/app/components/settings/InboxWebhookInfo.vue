<script setup lang="ts">
const props = defineProps<{ inboxId: string }>()
const config = useRuntimeConfig()
const toast = useToast()

const webhookUrl = computed(() =>
  `${config.public.apiBase}/webhook/whatsapp/${props.inboxId}`
)

async function copy() {
  await navigator.clipboard.writeText(webhookUrl.value)
  toast.add({ title: 'URL copiada', icon: 'i-lucide-check', color: 'success' })
}
</script>

<template>
  <UFormField
    label="URL do Webhook"
    description="Configure esta URL no painel Meta Business Manager → WhatsApp → Configuração. O token de verificação está nas configurações abaixo."
    class="flex max-sm:flex-col justify-between items-start gap-4"
  >
    <div class="flex items-center gap-2 w-full sm:w-auto">
      <code class="text-xs bg-elevated border border-default rounded px-2 py-1.5 font-mono truncate max-w-sm">
        {{ webhookUrl }}
      </code>
      <UButton
        icon="i-lucide-copy"
        size="xs"
        color="neutral"
        variant="ghost"
        @click="copy"
      />
    </div>
  </UFormField>
</template>
