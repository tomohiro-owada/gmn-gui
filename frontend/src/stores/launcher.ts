import { ref } from 'vue'
import { defineStore } from 'pinia'
import { ListRecentWorkDirs, DeleteSession } from '../../wailsjs/go/service/SessionService'
import { OpenProject, SelectDirectory } from '../../wailsjs/go/main/LauncherApp'
import type { service } from '../../wailsjs/go/models'

export const useLauncherStore = defineStore('launcher', () => {
  const projects = ref<service.WorkDirInfo[]>([])
  const loading = ref(false)

  async function fetchProjects() {
    loading.value = true
    try {
      projects.value = (await ListRecentWorkDirs()) ?? []
    } finally {
      loading.value = false
    }
  }

  async function openProject(dir: string, sessionID?: string) {
    await OpenProject(dir, sessionID ?? '')
  }

  async function selectNewProject() {
    const dir = await SelectDirectory()
    if (dir) {
      await openProject(dir)
    }
  }

  async function deleteProject(dir: string) {
    const project = projects.value.find(p => p.path === dir)
    if (!project?.sessions) return
    for (const s of project.sessions) {
      await DeleteSession(s.id)
    }
    await fetchProjects()
  }

  return { projects, loading, fetchProjects, openProject, selectNewProject, deleteProject }
})
