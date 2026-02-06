<script lang="ts" setup>
import { computed, ref } from 'vue'
import type { service } from '../../../wailsjs/go/models'
import MarkdownRenderer from './MarkdownRenderer.vue'

const props = defineProps<{
  message: service.ChatMessage
}>()

const isUser = computed(() => props.message.role === 'user')
const isToolCall = computed(() => props.message.role === 'tool_call')
const isToolResult = computed(() => props.message.role === 'tool_result')

// Collapse long user messages (> 4 lines or > 300 chars)
const LINE_THRESHOLD = 4
const CHAR_THRESHOLD = 300

const lines = computed(() => props.message.content.split('\n'))
const isLong = computed(() =>
  isUser.value && (lines.value.length > LINE_THRESHOLD || props.message.content.length > CHAR_THRESHOLD)
)
const expanded = ref(false)

const collapsedLabel = computed(() => {
  const n = lines.value.length
  const chars = props.message.content.length
  return `Pasted text (${n} lines, ${chars} chars)`
})
</script>

<template>
  <div class="flex" :class="isUser ? 'justify-end' : 'justify-start'">
    <!-- User message -->
    <div
      v-if="isUser"
      class="max-w-[70%] rounded-[18px] bg-primary text-primary-foreground px-4 py-2.5"
    >
      <!-- Collapsed long message -->
      <div v-if="isLong && !expanded">
        <button
          class="flex items-center gap-1.5 opacity-80 hover:opacity-100 transition-opacity"
          @click="expanded = true"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m9 18 6-6-6-6"/></svg>
          <span class="text-[0.85em]">{{ collapsedLabel }}</span>
        </button>
      </div>
      <!-- Expanded / short message -->
      <div v-else>
        <button
          v-if="isLong"
          class="flex items-center gap-1.5 opacity-80 hover:opacity-100 transition-opacity mb-1"
          @click="expanded = false"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>
          <span class="text-[0.85em]">{{ collapsedLabel }}</span>
        </button>
        <p class="whitespace-pre-wrap">{{ message.content }}</p>
      </div>
    </div>

    <!-- Model message -->
    <div
      v-else-if="message.role === 'model'"
      class="w-full"
    >
      <MarkdownRenderer :content="message.content" />
    </div>

    <!-- Tool call -->
    <div
      v-else-if="isToolCall"
      class="max-w-[80%] rounded-lg border border-border bg-card px-4 py-2.5 text-sm"
    >
      <div class="flex items-center gap-2 text-muted-foreground mb-1">
        <span class="text-xs font-mono">Tool Call</span>
      </div>
      <p class="font-mono text-xs font-semibold">{{ message.toolName }}</p>
      <pre
        v-if="message.toolArgs"
        class="mt-1 text-xs text-muted-foreground overflow-x-auto"
      >{{ message.toolArgs }}</pre>
    </div>

    <!-- Tool result -->
    <div
      v-else-if="isToolResult"
      class="max-w-[80%] rounded-lg border border-border bg-card px-4 py-2.5 text-sm"
    >
      <div class="flex items-center gap-2 text-muted-foreground mb-1">
        <span class="text-xs font-mono">Tool Result: {{ message.toolName }}</span>
      </div>
      <pre class="text-xs overflow-x-auto max-h-48 overflow-y-auto whitespace-pre-wrap">{{ message.content }}</pre>
    </div>
  </div>
</template>
