<script lang="ts" setup>
import { useSettingsStore } from '../stores/settings'
import { useChatStore } from '../stores/chat'
import { useI18n } from '../lib/i18n'
import type { Locale } from '../lib/i18n'

const settingsStore = useSettingsStore()
const chatStore = useChatStore()
const { t } = useI18n()
</script>

<template>
  <div class="flex-1 flex flex-col p-6 overflow-y-auto">
    <div class="flex items-center gap-3 mb-6">
      <router-link
        to="/"
        class="p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
      </router-link>
      <h2 class="text-xl font-bold">{{ t('settings.title') }}</h2>
    </div>

    <div class="space-y-4 max-w-md">
      <!-- Language -->
      <div>
        <label class="block text-sm font-medium mb-1.5">{{ t('settings.language') }}</label>
        <p class="text-xs text-muted-foreground mb-1.5">{{ t('settings.languageDesc') }}</p>
        <select
          :value="settingsStore.locale"
          class="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm
                 focus:outline-none focus:ring-2 focus:ring-ring"
          @change="settingsStore.changeLocale(($event.target as HTMLSelectElement).value as Locale)"
        >
          <option value="en">English</option>
          <option value="ja">日本語</option>
        </select>
      </div>

      <!-- Default Model -->
      <div>
        <label class="block text-sm font-medium mb-1.5">{{ t('settings.defaultModel') }}</label>
        <p class="text-xs text-muted-foreground mb-1.5">{{ t('settings.defaultModelDesc') }}</p>
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
        <label class="block text-sm font-medium mb-1.5">{{ t('settings.auth') }}</label>
        <div class="rounded-lg border border-border p-3 text-sm space-y-1">
          <div class="flex items-center gap-2">
            <span
              class="w-2 h-2 rounded-full"
              :class="settingsStore.authStatus?.authenticated ? 'bg-green-500' : 'bg-red-500'"
            />
            <span>{{ settingsStore.authStatus?.authenticated ? t('settings.authenticated') : t('settings.notAuthenticated') }}</span>
          </div>
          <p v-if="settingsStore.authStatus?.projectId" class="text-xs text-muted-foreground">
            {{ t('settings.project') }}: {{ settingsStore.authStatus.projectId }}
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
          {{ t('settings.reloadConfig') }}
        </button>
        <button
          class="w-full rounded-lg border border-destructive/50 text-destructive px-4 py-2 text-sm
                 hover:bg-destructive/10 transition-colors"
          @click="chatStore.clear()"
        >
          {{ t('settings.clearHistory') }}
        </button>
      </div>
    </div>
  </div>
</template>
