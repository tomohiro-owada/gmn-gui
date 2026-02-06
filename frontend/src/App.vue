<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { GetMode } from '../wailsjs/go/service/ModeService'
import { useChatStore } from './stores/chat'
import { useMCPStore } from './stores/mcp'
import { useSettingsStore } from './stores/settings'
import { useSessionStore } from './stores/session'
import { useI18n } from './lib/i18n'
import LauncherView from './views/LauncherView.vue'

const { t } = useI18n()
const route = useRoute()

const mode = ref('')
const chatStore = useChatStore()
const mcpStore = useMCPStore()
const settingsStore = useSettingsStore()
const sessionStore = useSessionStore()

onMounted(async () => {
  mode.value = await GetMode()

  if (mode.value === 'chat') {
    await initChatMode()
  }
  // LauncherView handles its own init
})

async function initChatMode() {
  chatStore.setupEvents()
  mcpStore.setupEvents()

  // Auto-save when streaming completes
  chatStore.setAutoSaveCallback(() => sessionStore.currentSessionId)

  await settingsStore.initialize()
  await chatStore.fetchSessionModel()
  await chatStore.fetchWorkDir()
  await mcpStore.fetchServers()

  // Load session ID from backend (set during startup)
  const { GetSessionID } = await import('../wailsjs/go/service/ModeService')
  const sid = await GetSessionID()
  if (sid) {
    sessionStore.currentSessionId = sid
    await chatStore.loadMessages()
  }
}
</script>

<template>
  <!-- Launcher mode -->
  <LauncherView v-if="mode === 'launcher'" />

  <!-- Chat mode: full-width, no sidebar -->
  <div v-else-if="mode === 'chat'" class="flex flex-col h-screen bg-background text-foreground">
    <main class="flex-1 flex flex-col overflow-hidden">
      <RouterView />
    </main>
  </div>
</template>
