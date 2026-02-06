<script lang="ts" setup>
import { onMounted } from 'vue'
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { useChatStore } from './stores/chat'
import { useMCPStore } from './stores/mcp'
import { useSettingsStore } from './stores/settings'

const route = useRoute()
const chatStore = useChatStore()
const mcpStore = useMCPStore()
const settingsStore = useSettingsStore()

onMounted(async () => {
  // Set up Wails event listeners
  chatStore.setupEvents()
  mcpStore.setupEvents()

  // Load initial data
  await settingsStore.initialize()
  await mcpStore.fetchServers()
})

const navItems = [
  { path: '/', label: 'Chat', icon: '💬' },
  { path: '/mcp', label: 'MCP', icon: '🔌' },
  { path: '/prompts', label: 'Prompts', icon: '📝' },
  { path: '/skills', label: 'Skills', icon: '⚡' },
  { path: '/settings', label: 'Settings', icon: '⚙️' },
]
</script>

<template>
  <div class="flex h-screen bg-background text-foreground dark">
    <!-- Sidebar -->
    <aside class="w-56 border-r border-border flex flex-col bg-card">
      <div class="p-4 border-b border-border">
        <h1 class="text-lg font-bold tracking-tight">gmn-gui</h1>
        <p class="text-xs text-muted-foreground mt-0.5">{{ settingsStore.model }}</p>
      </div>
      <nav class="flex-1 p-2 space-y-0.5">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
          :class="route.path === item.path ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
        >
          <span class="text-base">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>
      <div class="p-3 border-t border-border">
        <div class="flex items-center gap-2 text-xs">
          <span
            class="w-2 h-2 rounded-full"
            :class="settingsStore.authStatus?.authenticated ? 'bg-green-500' : 'bg-red-500'"
          />
          <span class="text-muted-foreground">
            {{ settingsStore.authStatus?.authenticated ? 'Connected' : 'Not authenticated' }}
          </span>
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <main class="flex-1 flex flex-col overflow-hidden">
      <RouterView />
    </main>
  </div>
</template>
