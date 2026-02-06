import { ref, computed } from 'vue'

export type Locale = 'en' | 'ja'

function detectSystemLocale(): Locale {
  const saved = localStorage.getItem('gmn-gui-locale') as Locale | null
  if (saved && (saved === 'en' || saved === 'ja')) {
    return saved
  }
  const navLang = navigator.language.toLowerCase()
  if (navLang.startsWith('ja')) {
    return 'ja'
  }
  return 'en'
}

const currentLocale = ref<Locale>(detectSystemLocale())

const messages: Record<Locale, Record<string, string>> = {
  en: {
    'sidebar.selectDir': 'Select directory...',
    'sidebar.newChat': 'New Chat',
    'sidebar.noConversations': 'No conversations yet',
    'sidebar.connected': 'Connected',
    'sidebar.notAuthenticated': 'Not authenticated',
    'chat.title': 'Chat',
    'chat.new': 'New',
    'chat.placeholder': 'Send a message...',
    'chat.send': 'Send',
    'chat.stop': 'Stop',
    'chat.emptyTitle': 'gmn-gui',
    'chat.emptySubtitle': 'Start a conversation with Gemini',
    'chat.emptyModel': 'Model',
    'chat.thinking': 'Thinking...',
    'settings.title': 'Settings',
    'settings.defaultModel': 'Default Model',
    'settings.defaultModelDesc': 'New chats will start with this model',
    'settings.language': 'Language',
    'settings.languageDesc': 'Display language for the interface',
    'settings.auth': 'Authentication',
    'settings.authenticated': 'Authenticated',
    'settings.notAuthenticated': 'Not authenticated',
    'settings.project': 'Project',
    'settings.reloadConfig': 'Reload Config',
    'settings.primaryColor': 'Accent Color',
    'settings.primaryColorDesc': 'Choose a color for buttons and highlights',
    'settings.clearHistory': 'Clear Chat History',
    'mcp.title': 'MCP Servers',
    'mcp.refresh': 'Refresh',
    'mcp.noServers': 'No MCP servers configured.',
    'mcp.addInSettings': 'Add servers in ~/.gemini/settings.json',
    'mcp.connect': 'Connect',
    'mcp.disconnect': 'Disconnect',
    'mcp.toolsAvailable': 'tools available',
    'launcher.title': 'Recent Projects',
    'launcher.newProject': 'Open Directory',
    'launcher.noProjects': 'No recent projects',
    'launcher.sessions': 'sessions',
    'launcher.open': 'Open',
    'launcher.login': 'Sign in with Google',
    'launcher.loggingIn': 'Signing in...',
    'launcher.logout': 'Sign out',
  },
  ja: {
    'sidebar.selectDir': 'ディレクトリを選択...',
    'sidebar.newChat': '新しいチャット',
    'sidebar.noConversations': 'まだ会話はありません',
    'sidebar.connected': '接続済み',
    'sidebar.notAuthenticated': '未認証',
    'chat.title': 'チャット',
    'chat.new': '新規',
    'chat.placeholder': 'メッセージを入力...',
    'chat.send': '送信',
    'chat.stop': '停止',
    'chat.emptyTitle': 'gmn-gui',
    'chat.emptySubtitle': 'Geminiと会話を始めましょう',
    'chat.emptyModel': 'モデル',
    'chat.thinking': '考え中...',
    'settings.title': '設定',
    'settings.defaultModel': 'デフォルトモデル',
    'settings.defaultModelDesc': '新しいチャットはこのモデルで開始されます',
    'settings.language': '言語',
    'settings.languageDesc': 'インターフェースの表示言語',
    'settings.auth': '認証',
    'settings.authenticated': '認証済み',
    'settings.notAuthenticated': '未認証',
    'settings.project': 'プロジェクト',
    'settings.reloadConfig': '設定を再読み込み',
    'settings.primaryColor': 'アクセントカラー',
    'settings.primaryColorDesc': 'ボタンやハイライトの色を選択',
    'settings.clearHistory': 'チャット履歴をクリア',
    'mcp.title': 'MCPサーバー',
    'mcp.refresh': '更新',
    'mcp.noServers': 'MCPサーバーが設定されていません。',
    'mcp.addInSettings': '~/.gemini/settings.json でサーバーを追加してください',
    'mcp.connect': '接続',
    'mcp.disconnect': '切断',
    'mcp.toolsAvailable': 'ツール利用可能',
    'launcher.title': '最近のプロジェクト',
    'launcher.newProject': 'ディレクトリを開く',
    'launcher.noProjects': 'プロジェクトがありません',
    'launcher.sessions': 'セッション',
    'launcher.open': '開く',
    'launcher.login': 'Googleでログイン',
    'launcher.loggingIn': 'ログイン中...',
    'launcher.logout': 'サインアウト',
  },
}

export function useI18n() {
  const locale = computed({
    get: () => currentLocale.value,
    set: (v: Locale) => { currentLocale.value = v },
  })

  function t(key: string): string {
    return messages[currentLocale.value]?.[key] ?? messages.en[key] ?? key
  }

  return { locale, t }
}

export function setLocale(l: Locale) {
  currentLocale.value = l
}

export function getLocale(): Locale {
  return currentLocale.value
}
