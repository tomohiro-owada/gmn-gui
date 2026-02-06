<script lang="ts" setup>
import { ref, watch, nextTick } from 'vue'
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
const textareaRef = ref<HTMLTextAreaElement | null>(null)

function autoResize() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
  // Line height ~1.5 * fontSize(~14px) = ~21px, 5 lines ≈ 105px + padding
  const maxHeight = 5 * 21 + 12
  el.style.height = Math.min(el.scrollHeight, maxHeight) + 'px'
}

watch(inputText, () => nextTick(autoResize))

function handleKeydown(e: KeyboardEvent) {
  if (e.isComposing || isComposing.value || e.keyCode === 229) return

  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}

function handleCompositionEnd() {
  setTimeout(() => { isComposing.value = false }, 50)
}

function send() {
  const text = inputText.value.trim()
  if (!text || props.disabled) return
  emit('send', text)
  inputText.value = ''
  nextTick(autoResize)
}
</script>

<template>
  <div class="border-t border-border p-4 bg-card">
    <div class="flex items-end gap-2 max-w-3xl mx-auto">
      <div class="flex-1 relative">
        <textarea
          ref="textareaRef"
          v-model="inputText"
          :disabled="disabled"
          :placeholder="t('chat.placeholder')"
          rows="1"
          class="w-full resize-none rounded-lg border border-input bg-background px-3 py-2.5
                 placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring
                 disabled:opacity-50 overflow-y-auto"
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
