import { defineStore } from 'pinia'
import { ref } from 'vue'
import { SendMessage, StopGeneration, ClearHistory, GetMessages, GetModel, SetModel, GetWorkDir, SetWorkDir, SubmitAskUserResponse, GetPlanMode, SetPlanMode } from '../../wailsjs/go/service/ChatService'
import { SaveCurrentSession } from '../../wailsjs/go/service/SessionService'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { service } from '../../wailsjs/go/models'

export interface StreamEvent {
  type: 'start' | 'content' | 'tool_call' | 'tool_result' | 'done' | 'error'
  text?: string
  toolName?: string
  toolArgs?: string
}

// Auto-save callback set by App.vue
let autoSaveSessionId: (() => string | null) | null = null

export const useChatStore = defineStore('chat', () => {
  const messages = ref<service.ChatMessage[]>([])
  const streamingText = ref('')
  const isStreaming = ref(false)
  const error = ref<string | null>(null)
  const sessionModel = ref('')
  const workDir = ref('')

  // ask_user dialog state
  const askUserVisible = ref(false)
  const askUserQuestions = ref<any[]>([])

  // plan mode
  const planMode = ref(false)

  function setAutoSaveCallback(cb: () => string | null) {
    autoSaveSessionId = cb
  }

  function setupEvents() {
    EventsOn('chat:stream', (event: StreamEvent) => {
      switch (event.type) {
        case 'start':
          isStreaming.value = true
          streamingText.value = ''
          error.value = null
          break
        case 'content':
          streamingText.value += event.text || ''
          break
        case 'tool_call':
          break
        case 'tool_result':
          break
        case 'done':
          isStreaming.value = false
          streamingText.value = ''
          // Auto-save session
          if (autoSaveSessionId) {
            const id = autoSaveSessionId()
            if (id) SaveCurrentSession(id).catch(() => {})
          }
          break
        case 'error':
          isStreaming.value = false
          error.value = event.text || 'Unknown error'
          streamingText.value = ''
          break
      }
    })

    EventsOn('chat:messages', (msgs: service.ChatMessage[]) => {
      messages.value = msgs ?? []
    })

    EventsOn('chat:ask_user', (questions: any[]) => {
      askUserQuestions.value = questions
      askUserVisible.value = true
    })
  }

  async function submitAskUserAnswer(answer: string) {
    askUserVisible.value = false
    askUserQuestions.value = []
    await SubmitAskUserResponse(answer)
  }

  async function fetchSessionModel() {
    sessionModel.value = await GetModel()
  }

  async function changeSessionModel(model: string) {
    await SetModel(model)
    sessionModel.value = model
  }

  async function fetchWorkDir() {
    workDir.value = await GetWorkDir()
  }

  async function changeWorkDir(dir: string) {
    await SetWorkDir(dir)
    workDir.value = dir
  }

  async function send(text: string) {
    if (!text.trim() || isStreaming.value) return
    error.value = null
    isStreaming.value = true
    streamingText.value = ''
    try {
      await SendMessage(text)
    } catch (e) {
      error.value = String(e)
      isStreaming.value = false
    }
  }

  async function stop() {
    await StopGeneration()
    isStreaming.value = false
  }

  async function clear() {
    await ClearHistory()
    messages.value = []
    streamingText.value = ''
    error.value = null
    sessionModel.value = ''
    workDir.value = ''
    await fetchSessionModel()
  }

  async function loadMessages() {
    messages.value = await GetMessages()
  }

  async function fetchPlanMode() {
    planMode.value = await GetPlanMode()
  }

  async function togglePlanMode() {
    const next = !planMode.value
    await SetPlanMode(next)
    planMode.value = next
  }

  return {
    messages,
    streamingText,
    isStreaming,
    error,
    sessionModel,
    workDir,
    askUserVisible,
    askUserQuestions,
    planMode,
    setupEvents,
    setAutoSaveCallback,
    fetchSessionModel,
    changeSessionModel,
    fetchWorkDir,
    changeWorkDir,
    send,
    stop,
    clear,
    loadMessages,
    submitAskUserAnswer,
    fetchPlanMode,
    togglePlanMode,
  }
})
