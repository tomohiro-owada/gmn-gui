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
            Model: {{ settingsStore.model }}
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
        class="flex items-center gap-2 text-muted-foreground text-sm"
      >
        <span class="animate-pulse">Thinking...</span>
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
