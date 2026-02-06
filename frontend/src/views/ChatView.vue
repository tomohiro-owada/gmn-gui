<script lang="ts" setup>
import { ref, computed, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useChatStore } from '../stores/chat'
import { useSettingsStore } from '../stores/settings'
import { useI18n } from '../lib/i18n'
import ChatInput from '../components/chat/ChatInput.vue'
import MessageBubble from '../components/chat/MessageBubble.vue'
import StreamingText from '../components/chat/StreamingText.vue'

const chatStore = useChatStore()
const settingsStore = useSettingsStore()
const route = useRoute()
const { t } = useI18n()

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

function shortenPath(path: string): string {
  const home = path.match(/^\/Users\/[^/]+/)
  if (home) {
    return path.replace(home[0], '~')
  }
  return path
}
</script>

<template>
  <div class="flex-1 flex flex-col h-full">
    <!-- Header: workDir + model selector + nav -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-border bg-card">
      <!-- Left: working directory -->
      <div class="flex items-center gap-2 text-sm text-muted-foreground min-w-0">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="shrink-0"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>
        <span class="truncate">{{ shortenPath(chatStore.workDir) }}</span>
      </div>

      <!-- Center: model selector -->
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

      <!-- Right: MCP + Settings icons -->
      <div class="flex items-center gap-1">
        <router-link
          to="/mcp"
          class="p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
          :class="route.path === '/mcp' ? 'bg-accent text-foreground' : ''"
          title="MCP Servers"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/></svg>
        </router-link>
        <router-link
          to="/settings"
          class="p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
          :class="route.path === '/settings' ? 'bg-accent text-foreground' : ''"
          title="Settings"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>
        </router-link>
      </div>
    </div>

    <!-- Messages area -->
    <div ref="messagesContainer" class="flex-1 overflow-y-auto p-4 space-y-4">
      <!-- Empty state -->
      <div
        v-if="chatStore.messages.length === 0 && !chatStore.isStreaming"
        class="flex-1 flex items-center justify-center h-full"
      >
        <p class="text-muted-foreground text-sm">
          {{ t('chat.emptySubtitle') }}
        </p>
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
