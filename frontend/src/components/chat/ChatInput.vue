<script lang="ts" setup>
import { ref } from 'vue'
import { useI18n } from '../../lib/i18n'

const { t } = useI18n()

const props = defineProps<{
  disabled: boolean
  isStreaming: boolean
}>()

const emit = defineEmits<{
  send: [text: string]
  stop: []
}>()

const inputText = ref('')
const isComposing = ref(false)

function handleKeydown(e: KeyboardEvent) {
  // Don't send during IME composition (e.g. Japanese input)
  // Check both native isComposing and our ref, plus keyCode 229 (composition)
  if (e.isComposing || isComposing.value || e.keyCode === 229) return

  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}

function handleCompositionEnd() {
  // Delay clearing so the final Enter keydown is still blocked
  setTimeout(() => { isComposing.value = false }, 50)
}

function send() {
  const text = inputText.value.trim()
  if (!text || props.disabled) return
  emit('send', text)
  inputText.value = ''
}
</script>

<template>
  <div class="border-t border-border p-4 bg-card">
    <div class="flex items-end gap-2 max-w-3xl mx-auto">
      <div class="flex-1 relative">
        <textarea
          v-model="inputText"
          :disabled="disabled"
          :placeholder="t('chat.placeholder')"
          rows="1"
          class="w-full resize-none rounded-lg border border-input bg-background px-3 py-2.5 text-sm
                 placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring
                 disabled:opacity-50 max-h-32 overflow-y-auto"
          @keydown="handleKeydown"
          @compositionstart="isComposing = true"
          @compositionend="handleCompositionEnd"
        />
      </div>
      <button
        v-if="isStreaming"
        class="shrink-0 rounded-lg bg-destructive px-4 py-2.5 text-sm font-medium text-destructive-foreground
               hover:bg-destructive/90 transition-colors"
        @click="emit('stop')"
      >
        {{ t('chat.stop') }}
      </button>
      <button
        v-else
        :disabled="!inputText.trim() || disabled"
        class="shrink-0 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground
               hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        @click="send"
      >
        {{ t('chat.send') }}
      </button>
    </div>
  </div>
</template>
