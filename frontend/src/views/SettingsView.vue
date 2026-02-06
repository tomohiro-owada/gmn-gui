<script lang="ts" setup>
import { useSettingsStore } from '../stores/settings'
import { useChatStore } from '../stores/chat'

const settingsStore = useSettingsStore()
const chatStore = useChatStore()
</script>

<template>
  <div class="flex-1 flex flex-col p-6 overflow-y-auto">
    <h2 class="text-xl font-bold mb-6">Settings</h2>

    <div class="space-y-4 max-w-md">
      <!-- Default Model -->
      <div>
        <label class="block text-sm font-medium mb-1.5">Default Model</label>
        <p class="text-xs text-muted-foreground mb-1.5">New chats will start with this model</p>
        <select
          :value="settingsStore.defaultModel"
          class="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm
                 focus:outline-none focus:ring-2 focus:ring-ring"
          @change="settingsStore.changeDefaultModel(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="m in settingsStore.availableModels" :key="m" :value="m">
            {{ m }}
          </option>
        </select>
      </div>

      <!-- Auth Status -->
      <div>
        <label class="block text-sm font-medium mb-1.5">Authentication</label>
        <div class="rounded-lg border border-border p-3 text-sm space-y-1">
          <div class="flex items-center gap-2">
            <span
              class="w-2 h-2 rounded-full"
              :class="settingsStore.authStatus?.authenticated ? 'bg-green-500' : 'bg-red-500'"
            />
            <span>{{ settingsStore.authStatus?.authenticated ? 'Authenticated' : 'Not authenticated' }}</span>
          </div>
          <p v-if="settingsStore.authStatus?.projectId" class="text-xs text-muted-foreground">
            Project: {{ settingsStore.authStatus.projectId }}
          </p>
          <p v-if="settingsStore.authStatus?.error" class="text-xs text-destructive">
            {{ settingsStore.authStatus.error }}
          </p>
        </div>
      </div>

      <!-- Actions -->
      <div class="space-y-2 pt-2">
        <button
          class="w-full rounded-lg border border-input px-4 py-2 text-sm hover:bg-accent transition-colors"
          @click="settingsStore.reloadConfig()"
        >
          Reload Config
        </button>
        <button
          class="w-full rounded-lg border border-destructive/50 text-destructive px-4 py-2 text-sm
                 hover:bg-destructive/10 transition-colors"
          @click="chatStore.clear()"
        >
          Clear Chat History
        </button>
      </div>
    </div>
  </div>
</template>
