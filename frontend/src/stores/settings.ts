import { defineStore } from 'pinia'
import { ref } from 'vue'
import { GetAuthStatus, GetDefaultModel, SetDefaultModel, AvailableModels, ReloadConfig } from '../../wailsjs/go/service/SettingsService'
import type { service } from '../../wailsjs/go/models'

export const useSettingsStore = defineStore('settings', () => {
  const authStatus = ref<service.AuthStatus | null>(null)
  const defaultModel = ref('gemini-2.5-flash')
  const availableModels = ref<string[]>([])
  const loading = ref(false)

  async function fetchAuthStatus() {
    try {
      authStatus.value = await GetAuthStatus()
    } catch (e) {
      authStatus.value = { authenticated: false, projectId: '', error: String(e) } as service.AuthStatus
    }
  }

  async function fetchDefaultModel() {
    defaultModel.value = await GetDefaultModel()
  }

  async function changeDefaultModel(newModel: string) {
    await SetDefaultModel(newModel)
    defaultModel.value = newModel
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
      fetchDefaultModel(),
      fetchAvailableModels(),
    ])
  }

  return {
    authStatus,
    defaultModel,
    availableModels,
    loading,
    fetchAuthStatus,
    changeDefaultModel,
    reloadConfig,
    initialize,
  }
})
