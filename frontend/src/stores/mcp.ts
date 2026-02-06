import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  ListServers,
  ConnectServer,
  DisconnectServer,
  AddServer,
  RemoveServer,
} from '../../wailsjs/go/service/MCPManager'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { service } from '../../wailsjs/go/models'

export const useMCPStore = defineStore('mcp', () => {
  const servers = ref<service.MCPServerStatus[]>([])
  const loading = ref(false)

  function setupEvents() {
    EventsOn('mcp:updated', (updated: service.MCPServerStatus[]) => {
      servers.value = updated ?? []
    })
  }

  async function fetchServers() {
    loading.value = true
    try {
      servers.value = (await ListServers()) ?? []
    } finally {
      loading.value = false
    }
  }

  async function connect(name: string) {
    loading.value = true
    try {
      await ConnectServer(name)
      await fetchServers()
    } finally {
      loading.value = false
    }
  }

  async function disconnect(name: string) {
    await DisconnectServer(name)
    await fetchServers()
  }

  async function addServer(name: string, command: string, args: string, env: string) {
    await AddServer(name, command, args, env)
    await fetchServers()
  }

  async function removeServer(name: string) {
    await RemoveServer(name)
    await fetchServers()
  }

  return {
    servers,
    loading,
    setupEvents,
    fetchServers,
    connect,
    disconnect,
    addServer,
    removeServer,
  }
})
