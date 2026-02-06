<script lang="ts" setup>
import { onMounted } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useChatStore } from './stores/chat'
import { useMCPStore } from './stores/mcp'
import { useSettingsStore } from './stores/settings'
import { useSessionStore } from './stores/session'
import { useI18n } from './lib/i18n'

const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const chatStore = useChatStore()
const mcpStore = useMCPStore()
const settingsStore = useSettingsStore()
const sessionStore = useSessionStore()

onMounted(async () => {
  chatStore.setupEvents()
  mcpStore.setupEvents()

  // Auto-save when streaming completes
  chatStore.setAutoSaveCallback(() => sessionStore.currentSessionId)

  await settingsStore.initialize()
  await chatStore.fetchSessionModel()
  await mcpStore.fetchServers()
  await sessionStore.fetchSessions()

  // Start with a new session
  if (!sessionStore.currentSessionId) {
    await sessionStore.startNew()
  }
})

async function handleNewChat() {
  // Save current session if it has messages
  if (chatStore.messages.length > 0 && sessionStore.currentSessionId) {
    await sessionStore.save()
  }
  await sessionStore.startNew()
  await chatStore.fetchSessionModel()
  await chatStore.fetchWorkDir()
  router.push('/')
}

async function handleSelectSession(id: string) {
  if (id === sessionStore.currentSessionId) {
    router.push('/')
    return
  }
  // Save current session first
  if (chatStore.messages.length > 0 && sessionStore.currentSessionId) {
    await sessionStore.save()
  }
  await sessionStore.load(id)
  await chatStore.loadMessages()
  await chatStore.fetchSessionModel()
  await chatStore.fetchWorkDir()
  router.push('/')
}

async function handleDeleteSession(id: string) {
  await sessionStore.remove(id)
  if (!sessionStore.currentSessionId) {
    await sessionStore.startNew()
    await chatStore.fetchSessionModel()
    await chatStore.fetchWorkDir()
  }
}

function formatDate(dateStr: string) {
  const d = new Date(dateStr)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' })
}
</script>

<template>
  <div class="flex h-screen bg-background text-foreground">
    <!-- Sidebar -->
    <aside class="w-60 border-r border-border flex flex-col bg-card">
      <!-- New Chat button + Directory selector -->
      <div class="p-3 border-b border-border space-y-2">
        <button
          class="w-full flex items-center gap-2 rounded-lg border border-border px-3 py-2 text-sm
                 hover:bg-accent transition-colors"
          @click="handleNewChat"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14M5 12h14"/></svg>
          {{ t('sidebar.newChat') }}
        </button>
      </div>

      <!-- Session list -->
      <div class="flex-1 overflow-y-auto p-2 space-y-0.5">
        <div
          v-for="s in sessionStore.sessions"
          :key="s.id"
          class="w-full text-left rounded-lg px-3 py-2 text-sm transition-colors group relative cursor-pointer"
          :class="s.id === sessionStore.currentSessionId
            ? 'bg-accent text-accent-foreground'
            : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'"
          @click="handleSelectSession(s.id)"
        >
          <p class="truncate text-xs font-medium pr-5">{{ s.title }}</p>
          <p class="text-[10px] text-muted-foreground mt-0.5">
            {{ s.model }} · {{ formatDate(s.updatedAt) }}
          </p>
          <!-- Delete button -->
          <button
            class="absolute right-2 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100
                   text-muted-foreground hover:text-destructive transition-all p-1"
            @click.stop="handleDeleteSession(s.id)"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18M6 6l12 12"/></svg>
          </button>
        </div>

        <!-- Empty state -->
        <p
          v-if="sessionStore.sessions.length === 0"
          class="text-xs text-muted-foreground text-center py-8"
        >
          {{ t('sidebar.noConversations') }}
        </p>
      </div>

      <!-- Bottom bar: status + nav -->
      <div class="p-3 border-t border-border flex items-center justify-between">
        <div class="flex items-center gap-2 text-xs">
          <span
            class="w-2 h-2 rounded-full"
            :class="settingsStore.authStatus?.authenticated ? 'bg-green-500' : 'bg-red-500'"
          />
          <span class="text-muted-foreground">
            {{ settingsStore.authStatus?.authenticated ? t('sidebar.connected') : t('sidebar.notAuthenticated') }}
          </span>
        </div>
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
    </aside>

    <!-- Main content -->
    <main class="flex-1 flex flex-col overflow-hidden">
      <RouterView />
    </main>
  </div>
</template>
