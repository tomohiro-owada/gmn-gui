<script lang="ts" setup>
import { useMCPStore } from '../stores/mcp'

const mcpStore = useMCPStore()
</script>

<template>
  <div class="flex-1 flex flex-col p-6 overflow-y-auto">
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-xl font-bold">MCP Servers</h2>
      <button
        class="rounded-lg border border-input px-3 py-1.5 text-sm hover:bg-accent transition-colors"
        :disabled="mcpStore.loading"
        @click="mcpStore.fetchServers()"
      >
        Refresh
      </button>
    </div>

    <!-- Empty state -->
    <div
      v-if="mcpStore.servers.length === 0"
      class="text-center text-muted-foreground py-12"
    >
      <p class="text-sm">No MCP servers configured.</p>
      <p class="text-xs mt-1">Add servers in ~/.gemini/settings.json</p>
    </div>

    <!-- Server list -->
    <div class="space-y-3">
      <div
        v-for="server in mcpStore.servers"
        :key="server.name"
        class="rounded-lg border border-border p-4"
      >
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-2">
            <span
              class="w-2 h-2 rounded-full"
              :class="server.connected ? 'bg-green-500' : 'bg-gray-500'"
            />
            <h3 class="font-medium text-sm">{{ server.name }}</h3>
          </div>
          <div class="flex gap-2">
            <button
              v-if="!server.connected"
              class="rounded px-3 py-1 text-xs bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              :disabled="mcpStore.loading"
              @click="mcpStore.connect(server.name)"
            >
              Connect
            </button>
            <button
              v-else
              class="rounded px-3 py-1 text-xs border border-input hover:bg-accent transition-colors"
              @click="mcpStore.disconnect(server.name)"
            >
              Disconnect
            </button>
          </div>
        </div>

        <p v-if="server.command" class="text-xs text-muted-foreground font-mono">
          {{ server.command }}
        </p>
        <p v-if="server.url" class="text-xs text-muted-foreground font-mono">
          {{ server.url }}
        </p>

        <div v-if="server.connected && server.toolCount > 0" class="mt-2">
          <p class="text-xs text-muted-foreground">{{ server.toolCount }} tools available</p>
          <div class="flex flex-wrap gap-1 mt-1">
            <span
              v-for="tool in server.tools"
              :key="tool"
              class="inline-block rounded bg-muted px-2 py-0.5 text-xs font-mono"
            >
              {{ tool }}
            </span>
          </div>
        </div>

        <p v-if="server.error" class="mt-2 text-xs text-destructive">
          {{ server.error }}
        </p>
      </div>
    </div>
  </div>
</template>
