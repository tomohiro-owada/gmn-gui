<script lang="ts" setup>
import { onMounted } from 'vue'
import { useLauncherStore } from '../stores/launcher'
import { useSettingsStore } from '../stores/settings'
import { useI18n } from '../lib/i18n'

const launcherStore = useLauncherStore()
const settingsStore = useSettingsStore()
const { t } = useI18n()

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
</script>

<template>
  <div class="flex flex-col h-screen bg-background text-foreground">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-border">
      <h1 class="text-sm font-semibold">{{ t('launcher.title') }}</h1>
      <div class="flex items-center gap-2">
        <span
          class="w-2 h-2 rounded-full"
          :class="settingsStore.authStatus?.authenticated ? 'bg-green-500' : 'bg-red-500'"
        />
      </div>
    </div>

    <!-- Project list -->
    <div class="flex-1 overflow-y-auto p-3 space-y-1.5">
      <div
        v-for="project in launcherStore.projects"
        :key="project.path"
        class="rounded-lg border border-border p-3 hover:bg-accent/50 transition-colors cursor-pointer group"
        @click="handleOpenProject(project.path, project.sessions?.[0]?.id)"
      >
        <div class="flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground shrink-0"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>
          <span class="text-sm font-medium truncate">{{ shortenPath(project.path) }}</span>
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
    <div class="p-3 border-t border-border">
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
