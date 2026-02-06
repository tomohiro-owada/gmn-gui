<script lang="ts" setup>
import { ref, computed, nextTick, watch } from 'vue'
import { useChatStore } from '../stores/chat'
import { useSettingsStore } from '../stores/settings'
import { useI18n } from '../lib/i18n'
import { SelectDirectory } from '../../wailsjs/go/main/App'
import ChatInput from '../components/chat/ChatInput.vue'
import MessageBubble from '../components/chat/MessageBubble.vue'
import StreamingText from '../components/chat/StreamingText.vue'

const chatStore = useChatStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()

async function handleSelectDir() {
  const dir = await SelectDirectory()
  if (dir) {
    await chatStore.changeWorkDir(dir)
  }
}
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
        <span class="text-sm font-medium">{{ t('chat.title') }}</span>
        <button
          class="text-xs text-muted-foreground hover:text-foreground transition-colors"
          @click="chatStore.clear()"
        >
          {{ t('chat.new') }}
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
        <div class="text-center space-y-4">
          <p class="text-muted-foreground text-sm">
            {{ t('chat.emptySubtitle') }}
          </p>
          <button
            class="inline-flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm transition-colors"
            :class="chatStore.workDir
              ? 'text-muted-foreground hover:bg-accent hover:text-foreground'
              : 'border border-dashed border-primary/40 text-primary bg-primary/5 hover:bg-primary/10'"
            @click="handleSelectDir"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>
            {{ chatStore.workDir || t('sidebar.selectDir') }}
          </button>
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
