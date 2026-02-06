<script lang="ts" setup>
import type { service } from '../../../wailsjs/go/models'

defineProps<{
  visible: boolean
  data: service.UsageResponse | null
}>()

const emit = defineEmits<{
  close: []
}>()

function formatResetTime(resetTime: string): string {
  const diff = new Date(resetTime).getTime() - Date.now()
  if (diff <= 0) return ''

  const totalMinutes = Math.ceil(diff / (1000 * 60))
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60

  if (hours > 0 && minutes > 0) return `Resets in ${hours}h ${minutes}m`
  if (hours > 0) return `Resets in ${hours}h`
  return `Resets in ${minutes}m`
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="emit('close')">
    <div class="bg-card border border-border rounded-xl shadow-lg max-w-lg w-full mx-4 p-5">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-sm font-semibold">Model Usage</h3>
        <button
          class="p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
          @click="emit('close')"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
        </button>
      </div>

      <div v-if="data?.error" class="text-sm text-destructive">{{ data.error }}</div>

      <div v-else-if="data?.buckets?.length">
        <!-- Table header -->
        <div class="grid grid-cols-[1fr_auto_auto] gap-x-4 text-xs font-semibold text-muted-foreground pb-2 border-b border-border">
          <span>Model</span>
          <span class="text-right w-16">Usage</span>
          <span class="text-right w-36">Reset</span>
        </div>

        <!-- Table rows -->
        <div
          v-for="bucket in data.buckets"
          :key="bucket.modelId"
          class="grid grid-cols-[1fr_auto_auto] gap-x-4 py-2 border-b border-border/50 text-sm"
        >
          <span class="truncate text-foreground">{{ bucket.modelId }}</span>
          <span class="text-right w-16 tabular-nums"
            :class="bucket.remainingFraction < 0.2 ? 'text-destructive' : bucket.remainingFraction < 0.5 ? 'text-amber-500' : 'text-muted-foreground'"
          >{{ (bucket.remainingFraction * 100).toFixed(1) }}%</span>
          <span class="text-right w-36 text-muted-foreground text-xs">
            {{ bucket.resetTime ? formatResetTime(bucket.resetTime) : '' }}
          </span>
        </div>
      </div>

      <div v-else class="text-sm text-muted-foreground">No usage data available.</div>

      <p class="text-[11px] text-muted-foreground mt-3">Usage limits span all sessions and reset daily.</p>
    </div>
  </div>
</template>
