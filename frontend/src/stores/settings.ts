import { defineStore } from 'pinia'
import { ref } from 'vue'
import { GetAuthStatus, GetDefaultModel, SetDefaultModel, AvailableModels, ReloadConfig, Login, Logout } from '../../wailsjs/go/service/SettingsService'
import type { service } from '../../wailsjs/go/models'
import { setLocale, getLocale, type Locale } from '../lib/i18n'

// Predefined primary color palette — soft pastel tones with dark text for readability
export const primaryColors = [
  { name: 'Slate',    hsl: '215 20% 65%',   fg: '220 30% 15%',  hex: '#94a3b8' },
  { name: 'Sky',      hsl: '199 70% 72%',   fg: '200 50% 15%',  hex: '#7dd3fc' },
  { name: 'Lavender', hsl: '250 50% 75%',   fg: '250 40% 20%',  hex: '#a5a0e4' },
  { name: 'Lilac',    hsl: '290 40% 72%',   fg: '290 30% 18%',  hex: '#c4a1d8' },
  { name: 'Rose',     hsl: '350 50% 75%',   fg: '350 40% 18%',  hex: '#e4a1ac' },
  { name: 'Peach',    hsl: '20 60% 75%',    fg: '20 45% 18%',   hex: '#e8b99a' },
  { name: 'Mint',     hsl: '160 40% 68%',   fg: '160 35% 15%',  hex: '#86ccb5' },
  { name: 'Sage',     hsl: '140 25% 68%',   fg: '140 25% 15%',  hex: '#96bba3' },
] as const

export type PrimaryColor = typeof primaryColors[number]

function applyPrimaryColor(color: PrimaryColor) {
  document.documentElement.style.setProperty('--primary', color.hsl)
  document.documentElement.style.setProperty('--primary-foreground', color.fg)
}

export const useSettingsStore = defineStore('settings', () => {
  const authStatus = ref<service.AuthStatus | null>(null)
  const defaultModel = ref('gemini-2.5-flash')
  const availableModels = ref<string[]>([])
  const locale = ref<Locale>('en')
  const loading = ref(false)
  const primaryColor = ref<string>(primaryColors[0].name)
  const fontSize = ref(16)

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

  async function login() {
    try {
      authStatus.value = await Login()
    } catch (e) {
      authStatus.value = { authenticated: false, projectId: '', error: String(e) } as service.AuthStatus
    }
  }

  async function logout() {
    authStatus.value = await Logout()
  }

  function changeLocale(newLocale: Locale) {
    locale.value = newLocale
    setLocale(newLocale)
    localStorage.setItem('gmn-gui-locale', newLocale)
  }

  function increaseFontSize() {
    if (fontSize.value < 22) {
      fontSize.value += 1
      localStorage.setItem('gmn-gui-font-size', String(fontSize.value))
    }
  }

  function decreaseFontSize() {
    if (fontSize.value > 10) {
      fontSize.value -= 1
      localStorage.setItem('gmn-gui-font-size', String(fontSize.value))
    }
  }

  function changePrimaryColor(name: string) {
    const color = primaryColors.find(c => c.name === name)
    if (!color) return
    primaryColor.value = name
    applyPrimaryColor(color)
    localStorage.setItem('gmn-gui-primary-color', name)
  }

  async function initialize() {
    // Sync locale from i18n (already auto-detected from localStorage or navigator.language)
    locale.value = getLocale()

    // Restore font size from localStorage
    const savedFontSize = localStorage.getItem('gmn-gui-font-size')
    if (savedFontSize) {
      const size = parseInt(savedFontSize, 10)
      if (size >= 10 && size <= 22) fontSize.value = size
    }

    // Restore primary color from localStorage
    const savedColor = localStorage.getItem('gmn-gui-primary-color')
    if (savedColor) {
      const color = primaryColors.find(c => c.name === savedColor)
      if (color) {
        primaryColor.value = savedColor
        applyPrimaryColor(color)
      }
    }

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
    locale,
    loading,
    primaryColor,
    fontSize,
    increaseFontSize,
    decreaseFontSize,
    fetchAuthStatus,
    changeDefaultModel,
    changeLocale,
    changePrimaryColor,
    reloadConfig,
    login,
    logout,
    initialize,
  }
})
