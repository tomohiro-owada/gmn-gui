<script lang="ts" setup>
import { ref, onMounted, watch } from 'vue'
import { useLauncherStore } from '../stores/launcher'
import { useSettingsStore, primaryColors } from '../stores/settings'
import { useI18n } from '../lib/i18n'
import type { Locale } from '../lib/i18n'
import { GetUsage } from '../../wailsjs/go/service/SettingsService'
import type { service } from '../../wailsjs/go/models'

const launcherStore = useLauncherStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()

const showSettings = ref(false)
const loginLoading = ref(false)
const usageData = ref<service.UsageResponse | null>(null)
const usageLoading = ref(false)

async function fetchUsage() {
  if (!settingsStore.authStatus?.authenticated) return
  usageLoading.value = true
  try {
    usageData.value = await GetUsage()
  } catch {
    usageData.value = null
  } finally {
    usageLoading.value = false
  }
}

function formatResetTime(resetTime: string): string {
  const diff = new Date(resetTime).getTime() - Date.now()
  if (diff <= 0) return ''
  const totalMinutes = Math.ceil(diff / (1000 * 60))
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60
  if (hours > 0 && minutes > 0) return `Resets in ${hours}h ${minutes}m`
  if (hours > 0) return `Resets in ${hours}h`
  return `Resets in ${minutes}m`
}

watch(showSettings, (v) => {
  if (v) fetchUsage()
})

onMounted(async () => {
  await settingsStore.initialize()
  await launcherStore.fetchProjects()
})

function shortenPath(path: string): string {
  const home = path.match(/^\/Users\/[^/]+/)
  if (home) {
    return path.replace(home[0], '~')
  }
  return path
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  const diffHr = Math.floor(diffMs / 3600000)
  const diffDay = Math.floor(diffMs / 86400000)

  if (diffMin < 1) return 'just now'
  if (diffMin < 60) return `${diffMin}m ago`
  if (diffHr < 24) return `${diffHr}h ago`
  if (diffDay < 7) return `${diffDay}d ago`
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

function handleOpenProject(dir: string, sessionID?: string) {
  launcherStore.openProject(dir, sessionID)
}

async function handleLogin() {
  loginLoading.value = true
  try {
    await settingsStore.login()
  } finally {
    loginLoading.value = false
  }
}
</script>

<template>
  <div class="flex flex-col h-screen bg-background text-foreground">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-border">
      <h1 class="text-sm font-semibold">{{ showSettings ? t('settings.title') : t('launcher.title') }}</h1>
      <div class="flex items-center gap-2">
        <span
          class="w-2 h-2 rounded-full"
          :class="settingsStore.authStatus?.authenticated ? 'bg-green-500' : 'bg-red-500'"
        />
        <button
          class="p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
          :class="showSettings ? 'bg-accent text-foreground' : ''"
          title="Settings"
          @click="showSettings = !showSettings"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>
        </button>
      </div>
    </div>

    <!-- Settings view (replaces project list) -->
    <div v-if="showSettings" class="flex-1 overflow-y-auto p-3 space-y-3">
      <!-- Auth -->
      <div>
        <label class="block text-xs font-medium text-muted-foreground mb-1">{{ t('settings.auth') }}</label>
        <div class="rounded-lg border border-border bg-background p-2.5 text-sm space-y-1.5">
          <div class="flex items-center gap-2">
            <span
              class="w-2 h-2 rounded-full"
              :class="settingsStore.authStatus?.authenticated ? 'bg-green-500' : 'bg-red-500'"
            />
            <span class="text-xs">{{ settingsStore.authStatus?.authenticated ? t('settings.authenticated') : t('settings.notAuthenticated') }}</span>
          </div>
          <p v-if="settingsStore.authStatus?.projectId" class="text-[11px] text-muted-foreground">
            {{ t('settings.project') }}: {{ settingsStore.authStatus.projectId }}
          </p>
          <p v-if="settingsStore.authStatus?.error" class="text-[11px] text-destructive">
            {{ settingsStore.authStatus.error }}
          </p>
          <button
            v-if="!settingsStore.authStatus?.authenticated"
            class="w-full mt-1 flex items-center justify-center gap-1.5 rounded-md border border-input
                   bg-background px-2.5 py-1.5 text-xs font-medium hover:bg-accent transition-colors"
            :disabled="loginLoading"
            @click="handleLogin"
          >
            <svg v-if="!loginLoading" xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" y1="12" x2="3" y2="12"/></svg>
            <span v-if="loginLoading" class="animate-spin w-3 h-3 border-2 border-current border-t-transparent rounded-full" />
            {{ loginLoading ? t('launcher.loggingIn') : t('launcher.login') }}
          </button>
          <button
            v-if="settingsStore.authStatus?.authenticated"
            class="w-full mt-1 rounded-md border border-input px-2.5 py-1.5 text-xs text-muted-foreground
                   hover:text-destructive hover:border-destructive/50 hover:bg-destructive/5 transition-colors"
            @click="settingsStore.logout()"
          >
            {{ t('launcher.logout') }}
          </button>
        </div>
      </div>

      <!-- Accent Color -->
      <div>
        <label class="block text-xs font-medium text-muted-foreground mb-1">{{ t('settings.primaryColor') }}</label>
        <div class="flex flex-wrap gap-1.5">
          <button
            v-for="color in primaryColors"
            :key="color.name"
            class="w-6 h-6 rounded-full border-2 transition-all hover:scale-110"
            :class="settingsStore.primaryColor === color.name ? 'border-foreground scale-110' : 'border-transparent'"
            :style="{ backgroundColor: color.hex }"
            :title="color.name"
            @click="settingsStore.changePrimaryColor(color.name)"
          />
        </div>
      </div>

      <!-- Language -->
      <div>
        <label class="block text-xs font-medium text-muted-foreground mb-1">{{ t('settings.language') }}</label>
        <select
          :value="settingsStore.locale"
          class="w-full rounded-md border border-input bg-background px-2 py-1 text-xs
                 focus:outline-none focus:ring-1 focus:ring-ring"
          @change="settingsStore.changeLocale(($event.target as HTMLSelectElement).value as Locale)"
        >
          <option value="en">English</option>
          <option value="ja">日本語</option>
        </select>
      </div>

      <!-- Default Model -->
      <div>
        <label class="block text-xs font-medium text-muted-foreground mb-1">{{ t('settings.defaultModel') }}</label>
        <select
          :value="settingsStore.defaultModel"
          class="w-full rounded-md border border-input bg-background px-2 py-1 text-xs
                 focus:outline-none focus:ring-1 focus:ring-ring"
          @change="settingsStore.changeDefaultModel(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="m in settingsStore.availableModels" :key="m" :value="m">
            {{ m }}
          </option>
        </select>
      </div>

      <!-- Usage -->
      <div v-if="settingsStore.authStatus?.authenticated">
        <div class="flex items-center justify-between mb-1">
          <label class="block text-xs font-medium text-muted-foreground">{{ t('settings.usage') }}</label>
          <button
            class="text-[11px] text-muted-foreground hover:text-foreground transition-colors"
            :disabled="usageLoading"
            @click="fetchUsage"
          >{{ usageLoading ? '...' : '↻' }}</button>
        </div>
        <div class="rounded-lg border border-border bg-background p-2 text-xs">
          <div v-if="usageData?.error" class="text-destructive">{{ usageData.error }}</div>
          <div v-else-if="usageData?.buckets?.length">
            <!-- Header -->
            <div class="grid grid-cols-[1fr_3rem_auto] gap-x-2 pb-1 border-b border-border/50 text-[11px] font-medium text-muted-foreground">
              <span>Model</span>
              <span class="text-right">Left</span>
              <span class="text-right">Reset</span>
            </div>
            <!-- Rows -->
            <div
              v-for="b in usageData.buckets"
              :key="b.modelId"
              class="grid grid-cols-[1fr_3rem_auto] gap-x-2 py-1 border-b border-border/30 last:border-0"
            >
              <span class="truncate">{{ b.modelId }}</span>
              <span class="text-right tabular-nums"
                :class="b.remainingFraction < 0.2 ? 'text-destructive' : b.remainingFraction < 0.5 ? 'text-amber-500' : ''"
              >{{ (b.remainingFraction * 100).toFixed(0) }}%</span>
              <span class="text-right text-muted-foreground text-[10px]">{{ b.resetTime ? formatResetTime(b.resetTime) : '' }}</span>
            </div>
          </div>
          <div v-else-if="usageLoading" class="text-muted-foreground text-center py-1">Loading...</div>
          <div v-else class="text-muted-foreground text-center py-1">No data</div>
        </div>
      </div>

      <!-- Reload Config -->
      <button
        class="w-full rounded-md border border-input px-2 py-1.5 text-xs hover:bg-accent transition-colors"
        @click="settingsStore.reloadConfig()"
      >
        {{ t('settings.reloadConfig') }}
      </button>
    </div>

    <!-- Project list -->
    <div v-else class="flex-1 overflow-y-auto p-3 space-y-1.5">
      <div
        v-for="project in launcherStore.projects"
        :key="project.path"
        class="rounded-lg border border-border p-3 hover:bg-accent/50 transition-colors cursor-pointer group"
        @click="handleOpenProject(project.path, project.sessions?.[0]?.id)"
      >
        <div class="flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground shrink-0"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>
          <span class="text-sm font-medium truncate flex-1">{{ shortenPath(project.path) }}</span>
          <button
            class="p-1 rounded-md text-muted-foreground/0 group-hover:text-muted-foreground
                   hover:!text-destructive hover:!bg-destructive/10 transition-colors shrink-0"
            title="Delete sessions"
            @click.stop="launcherStore.deleteProject(project.path)"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/></svg>
          </button>
        </div>
        <div class="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
          <span>{{ project.model }}</span>
          <span>·</span>
          <span>{{ formatDate(project.updatedAt) }}</span>
          <span>·</span>
          <span>{{ project.sessionCount }} {{ t('launcher.sessions') }}</span>
        </div>
      </div>

      <!-- Empty state -->
      <div
        v-if="launcherStore.projects.length === 0 && !launcherStore.loading"
        class="flex-1 flex items-center justify-center h-full"
      >
        <p class="text-sm text-muted-foreground">{{ t('launcher.noProjects') }}</p>
      </div>
    </div>

    <!-- Footer: New Project button -->
    <div v-if="!showSettings" class="p-3 border-t border-border">
      <button
        class="w-full flex items-center justify-center gap-2 rounded-lg border border-dashed border-primary/40
               bg-primary/5 px-3 py-2.5 text-sm text-primary hover:bg-primary/10 transition-colors"
        @click="launcherStore.selectNewProject()"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14M5 12h14"/></svg>
        {{ t('launcher.newProject') }}
      </button>
    </div>
  </div>
</template>
