import { defineStore } from 'pinia'
import { ref } from 'vue'
import { SendMessage, StopGeneration, ClearHistory, GetMessages, GetModel, SetModel, GetWorkDir, SetWorkDir, SubmitAskUserResponse, GetPlanMode, SetPlanMode } from '../../wailsjs/go/service/ChatService'
import { GetUsage } from '../../wailsjs/go/service/SettingsService'
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

  // usage dialog
  const usageVisible = ref(false)
  const usageData = ref<service.UsageResponse | null>(null)

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

  async function handleSlashCommand(text: string): Promise<boolean> {
    const cmd = text.trim().toLowerCase()
    if (cmd === '/usage' || cmd === '/stats') {
      try {
        const resp = await GetUsage()
        if (resp.error) {
          error.value = resp.error
        } else {
          usageData.value = resp
          usageVisible.value = true
        }
      } catch (e) {
        error.value = String(e)
      }
      return true
    }
    return false
  }

  // Detect MIME type from file extension
  function getMimeTypeFromExtension(filename: string): string {
    const ext = filename.toLowerCase().split('.').pop()
    const mimeMap: Record<string, string> = {
      // Text formats
      'md': 'text/markdown',
      'markdown': 'text/markdown',
      'txt': 'text/plain',
      'json': 'application/json',
      'xml': 'text/xml',
      'csv': 'text/csv',
      'yaml': 'text/yaml',
      'yml': 'text/yaml',
      'log': 'text/plain',
      'html': 'text/html',
      'htm': 'text/html',
      'css': 'text/css',
      'js': 'text/javascript',
      'ts': 'text/x-typescript',
      'py': 'text/x-python',
      'java': 'text/x-java',
      'c': 'text/x-c',
      'cpp': 'text/x-c++',
      'go': 'text/x-go',
      'rs': 'text/x-rust',
      // Images
      'jpg': 'image/jpeg',
      'jpeg': 'image/jpeg',
      'png': 'image/png',
      'gif': 'image/gif',
      'webp': 'image/webp',
      'bmp': 'image/bmp',
      'svg': 'image/svg+xml',
      // Documents
      'pdf': 'application/pdf',
      'doc': 'application/msword',
      'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      // Audio
      'mp3': 'audio/mp3',
      'wav': 'audio/wav',
      'ogg': 'audio/ogg',
      'flac': 'audio/flac',
      'aac': 'audio/aac',
      // Video
      'mp4': 'video/mp4',
      'webm': 'video/webm',
      'mov': 'video/mov',
      'avi': 'video/avi',
    }
    return mimeMap[ext || ''] || 'text/plain'
  }

  async function send(text: string) {
    return sendWithFiles(text, [])
  }

  async function sendWithFiles(text: string, files: File[]) {
    if ((!text.trim() && files.length === 0) || isStreaming.value) return
    error.value = null

    // Handle slash commands locally (only if no files attached)
    if (files.length === 0 && text.trim().startsWith('/')) {
      if (await handleSlashCommand(text)) return
    }

    isStreaming.value = true
    streamingText.value = ''
    try {
      if (files.length === 0) {
        await SendMessage(text)
      } else {
        // Save files to temp location and get paths
        const { SaveFilesToTemp } = await import('../../wailsjs/go/service/ChatService')

        // Read files as base64
        const fileData = await Promise.all(files.map(async (file) => {
          const buffer = await file.arrayBuffer()
          const bytes = new Uint8Array(buffer)
          const binary = Array.from(bytes).map(b => String.fromCharCode(b)).join('')
          const base64 = btoa(binary)

          // Use browser's MIME type if available and valid, otherwise detect from extension
          let mimeType = file.type
          if (!mimeType || mimeType === 'application/octet-stream' || mimeType === '') {
            mimeType = getMimeTypeFromExtension(file.name)
          }

          return {
            filename: file.name,
            mimeType: mimeType,
            data: base64
          }
        }))

        // Save files and get temp paths
        const attachedFiles = await SaveFilesToTemp(fileData)

        // Send message with file paths
        const { SendMessageWithFiles } = await import('../../wailsjs/go/service/ChatService')
        await SendMessageWithFiles(text, attachedFiles)
      }
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
    usageVisible,
    usageData,
    setupEvents,
    setAutoSaveCallback,
    fetchSessionModel,
    changeSessionModel,
    fetchWorkDir,
    changeWorkDir,
    send,
    sendWithFiles,
    stop,
    clear,
    loadMessages,
    submitAskUserAnswer,
    fetchPlanMode,
    togglePlanMode,
  }
})
