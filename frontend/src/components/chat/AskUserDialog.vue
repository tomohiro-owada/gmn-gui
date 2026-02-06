<script lang="ts" setup>
import { ref, watch } from 'vue'
import { SubmitAskUserResponse } from '../../../wailsjs/go/service/ChatService'

export interface AskUserQuestion {
  question: string
  header: string
  type: string // "choice" | "text" | "yesno"
  options?: { label: string; description: string }[]
}

const props = defineProps<{
  questions: AskUserQuestion[]
  visible: boolean
}>()

const emit = defineEmits<{
  done: []
}>()

const answers = ref<Record<number, string>>({})

watch(() => props.visible, (v) => {
  if (v) {
    const init: Record<number, string> = {}
    props.questions.forEach((_q, i) => { init[i] = '' })
    answers.value = init
  }
})

function selectOption(qIndex: number, label: string) {
  answers.value = { ...answers.value, [qIndex]: label }
}

function setAnswer(qIndex: number, value: string) {
  answers.value = { ...answers.value, [qIndex]: value }
}

async function submit() {
  const parts: string[] = []
  props.questions.forEach((q, i) => {
    parts.push(`${q.header}: ${answers.value[i] || '(no answer)'}`)
  })
  const answer = parts.join('\n')
  emit('done')
  await SubmitAskUserResponse(answer)
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-card border border-border rounded-xl shadow-lg max-w-md w-full mx-4 p-5">
      <h3 class="text-sm font-semibold text-muted-foreground mb-4">Question from AI</h3>

      <div v-for="(q, i) in questions" :key="i" class="mb-4">
        <label class="block text-sm font-medium mb-1.5">
          <span class="inline-block bg-muted text-muted-foreground text-xs px-1.5 py-0.5 rounded mr-1.5">{{ q.header }}</span>
          {{ q.question }}
        </label>

        <!-- Choice type -->
        <div v-if="q.type === 'choice' && q.options?.length" class="space-y-1.5 mt-2">
          <button
            v-for="opt in q.options"
            :key="opt.label"
            type="button"
            class="w-full text-left px-3 py-2 rounded-lg border text-sm transition-colors"
            :class="answers[i] === opt.label
              ? 'border-primary bg-primary/10 text-foreground'
              : 'border-border hover:border-primary/50 text-foreground'"
            @click="selectOption(i, opt.label)"
          >
            <span class="font-medium">{{ opt.label }}</span>
            <span v-if="opt.description" class="text-muted-foreground ml-1.5">- {{ opt.description }}</span>
          </button>
          <input
            :value="answers[i]"
            placeholder="Or type your own answer..."
            class="w-full mt-1.5 rounded-lg border border-input bg-background px-3 py-2 text-sm
                   placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            @input="setAnswer(i, ($event.target as HTMLInputElement).value)"
          />
        </div>

        <!-- Yes/No type -->
        <div v-else-if="q.type === 'yesno'" class="flex gap-2 mt-2">
          <button
            type="button"
            class="flex-1 px-3 py-2 rounded-lg border text-sm font-medium transition-colors"
            :class="answers[i] === 'Yes' ? 'border-primary bg-primary/10' : 'border-border hover:border-primary/50'"
            @click="selectOption(i, 'Yes')"
          >Yes</button>
          <button
            type="button"
            class="flex-1 px-3 py-2 rounded-lg border text-sm font-medium transition-colors"
            :class="answers[i] === 'No' ? 'border-primary bg-primary/10' : 'border-border hover:border-primary/50'"
            @click="selectOption(i, 'No')"
          >No</button>
        </div>

        <!-- Text type (default) -->
        <div v-else class="mt-2">
          <textarea
            :value="answers[i]"
            rows="2"
            class="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm
                   placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
            placeholder="Type your answer..."
            @input="setAnswer(i, ($event.target as HTMLTextAreaElement).value)"
          />
        </div>
      </div>

      <div class="flex justify-end gap-2 mt-4">
        <button
          type="button"
          class="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm font-medium
                 hover:bg-primary/90 transition-colors"
          @click="submit"
        >Submit</button>
      </div>
    </div>
  </div>
</template>
