<script lang="ts" setup>
import { computed } from 'vue'
import type { service } from '../../../wailsjs/go/models'
import MarkdownRenderer from './MarkdownRenderer.vue'

const props = defineProps<{
  message: service.ChatMessage
}>()

const isUser = computed(() => props.message.role === 'user')
const isToolCall = computed(() => props.message.role === 'tool_call')
const isToolResult = computed(() => props.message.role === 'tool_result')
</script>

<template>
  <div class="flex" :class="isUser ? 'justify-end' : 'justify-start'">
    <!-- User message -->
    <div
      v-if="isUser"
      class="max-w-[80%] rounded-lg bg-primary text-primary-foreground px-4 py-2.5 text-sm"
    >
      <p class="whitespace-pre-wrap">{{ message.content }}</p>
    </div>

    <!-- Model message -->
    <div
      v-else-if="message.role === 'model'"
      class="max-w-[80%] rounded-lg bg-muted px-4 py-2.5 text-sm"
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
