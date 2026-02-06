<script lang="ts" setup>
import { computed } from 'vue'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'

const props = defineProps<{
  content: string
}>()

function escapeHtml(str: string): string {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

const md: MarkdownIt = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
  highlight: (str: string, lang: string): string => {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return `<pre class="hljs"><code>${hljs.highlight(str, { language: lang }).value}</code></pre>`
      } catch { /* fallback */ }
    }
    return `<pre class="hljs"><code>${escapeHtml(str)}</code></pre>`
  },
})

const rendered = computed(() => md.render(props.content))
</script>

<template>
  <div class="markdown-body" v-html="rendered" />
</template>

<style>
.markdown-body {
  color: hsl(var(--foreground));
  line-height: 1.6;
  word-wrap: break-word;
}

.markdown-body pre {
  background-color: hsl(222.2 84% 4.9%);
  border-radius: 0.375rem;
  padding: 0.75rem;
  overflow-x: auto;
  margin: 0.5rem 0;
}

.markdown-body pre code {
  color: hsl(210 40% 98%);
}

.markdown-body code {
  font-size: 0.8rem;
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
}

.markdown-body :not(pre) > code {
  background-color: hsl(var(--muted));
  color: hsl(var(--foreground));
  padding: 0.15rem 0.35rem;
  border-radius: 0.25rem;
  font-size: 0.85em;
}

.markdown-body p {
  margin: 0.4rem 0;
}

.markdown-body ul, .markdown-body ol {
  margin: 0.4rem 0;
  padding-left: 1.5rem;
}

.markdown-body ul {
  list-style-type: disc;
}

.markdown-body ol {
  list-style-type: decimal;
}

.markdown-body li {
  margin: 0.15rem 0;
}

.markdown-body a {
  color: hsl(var(--primary));
  text-decoration: underline;
}

.markdown-body h1, .markdown-body h2, .markdown-body h3 {
  margin: 0.75rem 0 0.25rem;
  font-weight: 600;
}

.markdown-body h1 { font-size: 1.25em; }
.markdown-body h2 { font-size: 1.1em; }
.markdown-body h3 { font-size: 1em; }

.markdown-body blockquote {
  border-left: 3px solid hsl(var(--border));
  padding-left: 0.75rem;
  color: hsl(var(--muted-foreground));
  margin: 0.4rem 0;
}

.markdown-body table {
  border-collapse: collapse;
  width: 100%;
  margin: 0.5rem 0;
}

.markdown-body th, .markdown-body td {
  border: 1px solid hsl(var(--border));
  padding: 0.25rem 0.5rem;
  text-align: left;
}

.markdown-body th {
  background-color: hsl(var(--muted));
  font-weight: 600;
}

.markdown-body hr {
  border: none;
  border-top: 1px solid hsl(var(--border));
  margin: 0.75rem 0;
}

.markdown-body strong {
  font-weight: 600;
}
</style>
