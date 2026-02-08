<script lang="ts" setup>
import { ref, watch, nextTick, computed } from 'vue'
import { useI18n } from '../../lib/i18n'
import { Paperclip, X } from 'lucide-vue-next'

const { t } = useI18n()

const props = defineProps<{
  disabled: boolean
  isStreaming: boolean
}>()

const emit = defineEmits<{
  send: [data: { text: string; files: File[] }]
  stop: []
}>()

const inputText = ref('')
const isComposing = ref(false)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const attachedFiles = ref<File[]>([])
const isDragging = ref(false)

function autoResize() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
  // Line height ~1.5 * fontSize(~14px) = ~21px, 5 lines ≈ 105px + padding
  const maxHeight = 5 * 21 + 12
  el.style.height = Math.min(el.scrollHeight, maxHeight) + 'px'
}

watch(inputText, () => nextTick(autoResize))

// Drag and drop handlers
function handleDragEnter(e: DragEvent) {
  e.preventDefault()
  e.stopPropagation()
  if (e.dataTransfer?.types.includes('Files')) {
    isDragging.value = true
  }
}

function handleDragOver(e: DragEvent) {
  e.preventDefault()
  e.stopPropagation()
}

function handleDragLeave(e: DragEvent) {
  e.preventDefault()
  e.stopPropagation()
  // Only hide overlay if leaving the component entirely
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  if (
    e.clientX <= rect.left ||
    e.clientX >= rect.right ||
    e.clientY <= rect.top ||
    e.clientY >= rect.bottom
  ) {
    isDragging.value = false
  }
}

function handleDrop(e: DragEvent) {
  e.preventDefault()
  e.stopPropagation()
  isDragging.value = false

  if (e.dataTransfer?.files) {
    attachedFiles.value.push(...Array.from(e.dataTransfer.files))
  }
}

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

function openFilePicker() {
  fileInputRef.value?.click()
}

function handleFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files) {
    attachedFiles.value.push(...Array.from(input.files))
    input.value = '' // Reset input
  }
}

function removeFile(index: number) {
  attachedFiles.value.splice(index, 1)
}

function send() {
  const text = inputText.value.trim()
  const files = attachedFiles.value
  if ((!text && files.length === 0) || props.disabled) return

  emit('send', { text, files })
  inputText.value = ''
  attachedFiles.value = []
  nextTick(autoResize)
}

const canSend = computed(() => {
  return (inputText.value.trim() || attachedFiles.value.length > 0) && !props.disabled
})
</script>

<template>
  <div
    class="border-t border-border p-4 bg-card relative"
    @dragenter="handleDragEnter"
    @dragover="handleDragOver"
    @dragleave="handleDragLeave"
    @drop="handleDrop"
  >
    <!-- Drag overlay -->
    <div
      v-if="isDragging"
      class="absolute inset-0 bg-primary/10 border-2 border-dashed border-primary z-50
             flex items-center justify-center pointer-events-none"
    >
      <div class="text-center">
        <div class="text-2xl mb-2">📎</div>
        <div class="text-sm font-medium text-primary">Drop files here</div>
      </div>
    </div>

    <div class="max-w-3xl mx-auto">
      <!-- Attached files preview -->
      <div v-if="attachedFiles.length > 0" class="mb-3 flex flex-wrap gap-2">
        <div
          v-for="(file, index) in attachedFiles"
          :key="index"
          class="flex items-center gap-2 px-3 py-1.5 bg-muted rounded-lg text-sm"
        >
          <span class="truncate max-w-[200px]">{{ file.name }}</span>
          <button
            @click="removeFile(index)"
            class="shrink-0 hover:bg-muted-foreground/20 rounded p-0.5"
          >
            <X :size="14" />
          </button>
        </div>
      </div>

      <div class="flex items-end gap-2">
        <!-- File attach button -->
        <button
          :disabled="disabled"
          class="shrink-0 rounded-lg p-2.5 hover:bg-accent transition-colors disabled:opacity-50"
          @click="openFilePicker"
          title="Attach file"
        >
          <Paperclip :size="20" />
        </button>
        <input
          ref="fileInputRef"
          type="file"
          multiple
          accept="image/*,text/*,.md,.markdown,.yaml,.yml,.json,.xml,.csv,.log,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,audio/*,video/*"
          class="hidden"
          @change="handleFileSelect"
        />

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
          :disabled="!canSend"
          class="shrink-0 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground
                 hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          @click="send"
        >
          {{ t('chat.send') }}
        </button>
      </div>
    </div>
  </div>
</template>
