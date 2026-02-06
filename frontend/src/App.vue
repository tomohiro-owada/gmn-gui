<script lang="ts" setup>
import { RouterView } from 'vue-router'
</script>

<template>
  <div class="flex h-screen bg-background text-foreground">
    <!-- Sidebar -->
    <aside class="w-64 border-r border-border flex flex-col">
      <div class="p-4 border-b border-border">
        <h1 class="text-lg font-bold">gmn-gui</h1>
      </div>
      <nav class="flex-1 p-2 space-y-1">
        <router-link
          to="/"
          class="flex items-center px-3 py-2 rounded-md text-sm hover:bg-accent"
          active-class="bg-accent"
        >
          Chat
        </router-link>
        <router-link
          to="/mcp"
          class="flex items-center px-3 py-2 rounded-md text-sm hover:bg-accent"
          active-class="bg-accent"
        >
          MCP
        </router-link>
        <router-link
          to="/prompts"
          class="flex items-center px-3 py-2 rounded-md text-sm hover:bg-accent"
          active-class="bg-accent"
        >
          Prompts
        </router-link>
        <router-link
          to="/skills"
          class="flex items-center px-3 py-2 rounded-md text-sm hover:bg-accent"
          active-class="bg-accent"
        >
          Skills
        </router-link>
        <router-link
          to="/settings"
          class="flex items-center px-3 py-2 rounded-md text-sm hover:bg-accent"
          active-class="bg-accent"
        >
          Settings
        </router-link>
      </nav>
    </aside>

    <!-- Main content -->
    <main class="flex-1 flex flex-col overflow-hidden">
      <RouterView />
    </main>
  </div>
</template>
