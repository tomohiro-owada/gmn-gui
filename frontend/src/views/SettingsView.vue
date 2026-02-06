<script lang="ts" setup>
import { Greet } from '../../wailsjs/go/main/App'
import { ref } from 'vue'

const greeting = ref('')

async function testConnection() {
  greeting.value = await Greet('gmn-gui')
}
</script>

<template>
  <div class="flex-1 flex flex-col p-4">
    <h2 class="text-xl font-bold mb-4">Settings</h2>
    <button
      class="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:opacity-90"
      @click="testConnection"
    >
      Test Wails Connection
    </button>
    <p v-if="greeting" class="mt-4">{{ greeting }}</p>
  </div>
</template>
