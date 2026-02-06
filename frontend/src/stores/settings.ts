import { defineStore } from 'pinia'
import { ref, onMounted } from 'vue'
import { GetAuthStatus, GetModel, SetModel, AvailableModels, ReloadConfig } from '../../wailsjs/go/service/SettingsService'
import type { service } from '../../wailsjs/go/models'

export const useSettingsStore = defineStore('settings', () => {
  const authStatus = ref<service.AuthStatus | null>(null)
  const model = ref('gemini-2.5-flash')
  const availableModels = ref<string[]>([])
  const loading = ref(false)

  async function fetchAuthStatus() {
    try {
      authStatus.value = await GetAuthStatus()
    } catch (e) {
      authStatus.value = { authenticated: false, projectId: '', error: String(e) } as service.AuthStatus
    }
  }

  async function fetchModel() {
    model.value = await GetModel()
  }

  async function changeModel(newModel: string) {
    await SetModel(newModel)
    model.value = newModel
  }

  async function fetchAvailableModels() {
    availableModels.value = await AvailableModels()
  }

  async function reloadConfig() {
    loading.value = true
    try {
      await ReloadConfig()
      await fetchAuthStatus()
    } finally {
      loading.value = false
    }
  }

  async function initialize() {
    await Promise.all([
      fetchAuthStatus(),
      fetchModel(),
      fetchAvailableModels(),
    ])
  }

  return {
    authStatus,
    model,
    availableModels,
    loading,
    fetchAuthStatus,
    changeModel,
    reloadConfig,
    initialize,
  }
})
