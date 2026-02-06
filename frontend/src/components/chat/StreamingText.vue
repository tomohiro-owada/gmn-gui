<script lang="ts" setup>
import MarkdownRenderer from './MarkdownRenderer.vue'

defineProps<{
  text: string
}>()
</script>

<template>
  <div class="flex justify-start">
    <div class="w-full">
      <MarkdownRenderer :content="text" />
      <span class="inline-block w-1.5 h-4 bg-foreground/60 animate-pulse ml-0.5" />
    </div>
  </div>
</template>
