<script lang="ts" setup>
import { ref, computed, nextTick, watch } from 'vue'
import { useChatStore } from '../stores/chat'
import { useSettingsStore } from '../stores/settings'
import ChatInput from '../components/chat/ChatInput.vue'
import MessageBubble from '../components/chat/MessageBubble.vue'
import StreamingText from '../components/chat/StreamingText.vue'

const chatStore = useChatStore()
const settingsStore = useSettingsStore()
const messagesContainer = ref<HTMLElement | null>(null)

const showStreamingBubble = computed(() => {
  return chatStore.isStreaming && chatStore.streamingText.length > 0
})

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

watch(() => chatStore.messages.length, scrollToBottom)
watch(() => chatStore.streamingText, scrollToBottom)

async function handleSend(text: string) {
  await chatStore.send(text)
  scrollToBottom()
}
</script>

<template>
  <div class="flex-1 flex flex-col h-full">
    <!-- Header with session model selector -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-border bg-card">
      <div class="flex items-center gap-2">
        <span class="text-sm font-medium">Chat</span>
        <button
          class="text-xs text-muted-foreground hover:text-foreground transition-colors"
          @click="chatStore.clear()"
        >
          New
        </button>
      </div>
      <select
        :value="chatStore.sessionModel"
        class="rounded border border-input bg-background px-2 py-1 text-xs
               focus:outline-none focus:ring-1 focus:ring-ring"
        @change="chatStore.changeSessionModel(($event.target as HTMLSelectElement).value)"
      >
        <option v-for="m in settingsStore.availableModels" :key="m" :value="m">
          {{ m }}
        </option>
      </select>
    </div>

    <!-- Messages area -->
    <div ref="messagesContainer" class="flex-1 overflow-y-auto p-4 space-y-4">
      <!-- Empty state -->
      <div
        v-if="chatStore.messages.length === 0 && !chatStore.isStreaming"
        class="flex-1 flex items-center justify-center h-full"
      >
        <div class="text-center space-y-3">
          <p class="text-2xl font-semibold">gmn-gui</p>
          <p class="text-muted-foreground text-sm">
            Start a conversation with Gemini
          </p>
          <p class="text-xs text-muted-foreground">
            Model: {{ chatStore.sessionModel }}
          </p>
        </div>
      </div>

      <!-- Message list -->
      <MessageBubble
        v-for="msg in chatStore.messages"
        :key="msg.id"
        :message="msg"
      />

      <!-- Streaming response -->
      <StreamingText
        v-if="showStreamingBubble"
        :text="chatStore.streamingText"
      />

      <!-- Streaming indicator (no text yet) -->
      <div
        v-if="chatStore.isStreaming && chatStore.streamingText.length === 0"
        class="flex justify-start"
      >
        <div class="rounded-lg bg-muted px-4 py-3 flex items-center gap-1.5">
          <span class="thinking-dot w-2 h-2 rounded-full bg-foreground/40" style="animation-delay: 0ms" />
          <span class="thinking-dot w-2 h-2 rounded-full bg-foreground/40" style="animation-delay: 150ms" />
          <span class="thinking-dot w-2 h-2 rounded-full bg-foreground/40" style="animation-delay: 300ms" />
        </div>
      </div>

      <!-- Error display -->
      <div
        v-if="chatStore.error"
        class="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-sm text-destructive"
      >
        {{ chatStore.error }}
      </div>
    </div>

    <!-- Input area -->
    <ChatInput
      :disabled="chatStore.isStreaming"
      :is-streaming="chatStore.isStreaming"
      @send="handleSend"
      @stop="chatStore.stop"
    />
  </div>
</template>
