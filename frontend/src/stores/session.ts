import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  ListSessions,
  SaveCurrentSession,
  LoadSession,
  DeleteSession,
  NewSession,
} from '../../wailsjs/go/service/SessionService'
import type { service } from '../../wailsjs/go/models'

export const useSessionStore = defineStore('session', () => {
  const sessions = ref<service.SessionSummary[]>([])
  const currentSessionId = ref<string | null>(null)

  async function fetchSessions() {
    sessions.value = (await ListSessions()) ?? []
  }

  async function save() {
    if (!currentSessionId.value) return
    await SaveCurrentSession(currentSessionId.value)
    await fetchSessions()
  }

  async function load(id: string) {
    await LoadSession(id)
    currentSessionId.value = id
  }

  async function remove(id: string) {
    await DeleteSession(id)
    if (currentSessionId.value === id) {
      currentSessionId.value = null
    }
    await fetchSessions()
  }

  async function startNew() {
    const id = await NewSession()
    currentSessionId.value = id
    await fetchSessions()
    return id
  }

  return {
    sessions,
    currentSessionId,
    fetchSessions,
    save,
    load,
    remove,
    startNew,
  }
})
