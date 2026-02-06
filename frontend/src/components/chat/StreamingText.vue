<script lang="ts" setup>
import MarkdownRenderer from './MarkdownRenderer.vue'

defineProps<{
  text: string
}>()
</script>

<template>
  <div class="flex justify-start">
    <div class="max-w-[80%] rounded-lg bg-muted px-4 py-2.5 text-sm">
      <MarkdownRenderer :content="text" />
      <span class="inline-block w-1.5 h-4 bg-foreground/60 animate-pulse ml-0.5" />
    </div>
  </div>
</template>
