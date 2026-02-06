import { defineStore } from 'pinia'
import { ref } from 'vue'
import { SendMessage, StopGeneration, ClearHistory, GetMessages, GetModel, SetModel } from '../../wailsjs/go/service/ChatService'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { service } from '../../wailsjs/go/models'

export interface StreamEvent {
  type: 'start' | 'content' | 'tool_call' | 'tool_result' | 'done' | 'error'
  text?: string
  toolName?: string
  toolArgs?: string
}

export const useChatStore = defineStore('chat', () => {
  const messages = ref<service.ChatMessage[]>([])
  const streamingText = ref('')
  const isStreaming = ref(false)
  const error = ref<string | null>(null)
  const sessionModel = ref('')

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
          break
        case 'error':
          isStreaming.value = false
          error.value = event.text || 'Unknown error'
          streamingText.value = ''
          break
      }
    })

    EventsOn('chat:messages', (msgs: service.ChatMessage[]) => {
      messages.value = msgs
    })
  }

  async function fetchSessionModel() {
    sessionModel.value = await GetModel()
  }

  async function changeSessionModel(model: string) {
    await SetModel(model)
    sessionModel.value = model
  }

  async function send(text: string) {
    if (!text.trim() || isStreaming.value) return
    error.value = null
    try {
      await SendMessage(text)
    } catch (e) {
      error.value = String(e)
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
    await fetchSessionModel()
  }

  async function loadMessages() {
    messages.value = await GetMessages()
  }

  return {
    messages,
    streamingText,
    isStreaming,
    error,
    sessionModel,
    setupEvents,
    fetchSessionModel,
    changeSessionModel,
    send,
    stop,
    clear,
    loadMessages,
  }
})
