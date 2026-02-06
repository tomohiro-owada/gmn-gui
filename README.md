# gmn-gui

[English](#english) | [日本語](#日本語)

<img width="1046" height="752" alt="Screenshot 2026-02-06 at 3 01 11 PM" src="https://github.com/user-attachments/assets/c38c231f-c8c6-41ce-9139-fa17d06bb473" />

---

<a id="日本語"></a>

## 日本語

**gmn-gui** は [Gemini CLI](https://github.com/GoogleCloudPlatform/gemini-cli) の非公式 GUI クライアントです。Google の Gemini Code Assist API を利用し、AI によるコーディング支援をネイティブデスクトップアプリとして提供します。



### 特徴

- **プロジェクトランチャー** — 最近のプロジェクトをワンクリックで開き、過去のセッションを復元
- **AI チャット** — Gemini モデルとのリアルタイムストリーミング会話
- **15 種類の組み込みツール** — ファイル操作、シェル実行、Web 検索、grep など
- **Plan Mode** — 読み取り専用ツールのみに制限し、安全にコードを調査
- **MCP サーバー対応** — Model Context Protocol でツールを拡張
- **セッション管理** — 会話を自動保存・復元、プロジェクトごとに整理
- **多言語対応** — 日本語 / 英語 UI（システム言語を自動検出）
- **テーマカスタマイズ** — 8 色のアクセントカラー、フォントサイズ調整
- **API 使用量表示** — モデルごとの残りクォータをリアルタイム確認
- **Markdown レンダリング** — シンタックスハイライト付きコードブロック

### 動作要件

- macOS 12 以降（Apple Silicon）
- [Gemini Code Assist](https://codeassist.google/) のアカウント（Google アカウントで無料利用可能）

### インストール

#### リリースビルドを使う場合

1. [Releases](https://github.com/tomohiro-owada/gmn-gui/releases) から `gmn-gui-darwin-arm64.zip` をダウンロード
2. ZIP を展開して `gmn-gui.app` を `/Applications` に移動

#### macOS Gatekeeper の警告について

gmn-gui は署名されていない野良アプリのため、初回起動時に「"gmn-gui" は開発元が未確認のため開けません」と表示されます。

**回避方法：**

```bash
# ターミナルで以下を実行（1回だけ）
xattr -cr /Applications/gmn-gui.app
```

または：

1. Finder で `gmn-gui.app` を **右クリック**（Control + クリック）→「開く」を選択
2. 警告ダイアログで「開く」をクリック

> 2 回目以降は通常通り起動できます。

#### ソースからビルドする場合

```bash
# 前提条件
# - Go 1.23 以降
# - Node.js 18 以降
# - Wails CLI v2

go install github.com/wailsapp/wails/v2/cmd/wails@latest

git clone https://github.com/tomohiro-owada/gmn-gui.git
cd gmn-gui
wails build
```

ビルド後、`build/bin/gmn-gui.app` が生成されます。

### 使い方

#### 1. 初回ログイン

1. アプリを起動すると**ランチャー**が表示されます
2. 右上の ⚙ アイコンをクリックして設定画面を開く
3. 「Google でログイン」をクリックしてブラウザで認証
4. 認証完了後、自動的にプロジェクト ID が取得されます

#### 2. プロジェクトを開く

1. 「ディレクトリを開く」ボタンでプロジェクトフォルダを選択
2. チャットウィンドウが新しいウィンドウで開きます
3. 次回以降はランチャーの一覧からワンクリックで開けます

#### 3. AI とチャット

- テキスト入力欄にメッセージを入力し、Enter で送信（Shift+Enter で改行）
- AI はファイルの読み書き、シェルコマンド実行、Web 検索などのツールを自動的に使用します
- 右上のモデルセレクターで使用モデルを切り替え可能

#### 4. Plan Mode

- チャットヘッダーの「Plan」ボタンで Plan Mode を ON/OFF
- ON にすると AI は読み取り専用のツール（ファイル読み込み、grep、ディレクトリ一覧など）のみ使用可能
- コードの調査や分析を安全に行いたい場合に便利

#### 5. MCP サーバー

- サイドバーの MCP タブで接続中のサーバーを管理
- `~/.gemini/settings.json` にサーバーを設定：

```json
{
  "mcpServers": {
    "my-server": {
      "command": "npx",
      "args": ["-y", "@example/mcp-server"]
    }
  }
}
```

#### 6. 設定

| 項目 | 説明 |
|------|------|
| アクセントカラー | 8 色から UI のテーマカラーを選択 |
| 言語 | 日本語 / English |
| デフォルトモデル | 新規チャットで使用するモデル |
| 使用量 | モデルごとの残りクォータとリセット時間 |

#### 7. スラッシュコマンド

| コマンド | 説明 |
|---------|------|
| `/usage` または `/stats` | モデルの使用量・残りクォータを表示 |

### 対応モデル

| モデル | 説明 |
|-------|------|
| gemini-3-pro-preview | 最高精度（プレビュー） |
| gemini-3-flash-preview | 高速・高精度（プレビュー） |
| gemini-2.5-pro | 高精度 |
| gemini-2.5-flash | 標準（デフォルト） |
| gemini-2.5-flash-lite | 軽量・最速 |

### 組み込みツール

| ツール | 説明 |
|-------|------|
| `run_shell_command` | シェルコマンドの実行 |
| `read_file` | ファイルの読み取り（行範囲指定可） |
| `write_file` | ファイルの作成・上書き |
| `replace` | ファイル内のテキスト置換 |
| `read_many_files` | glob パターンで複数ファイル読み取り |
| `list_directory` | ディレクトリ一覧 |
| `glob` | パターンによるファイル検索 |
| `grep_search` | 正規表現によるファイル内検索 |
| `google_web_search` | Web 検索 |
| `web_fetch` | URL のコンテンツ取得 |
| `ask_user` | ユーザーへの質問（選択肢/テキスト） |
| `write_todos` | GEMINI.md にタスク保存 |
| `save_memory` | GEMINI.md にメモ保存 |

### 設定ファイル

```
~/.gemini/
├── settings.json          # グローバル設定・MCP サーバー
├── oauth_creds.json       # 認証情報（自動生成）
├── gmn_state.json         # キャッシュ状態
└── gmn-gui/
    └── sessions/          # チャットセッション
```

### ライセンス

Apache License 2.0

一部コードは [Gemini CLI](https://github.com/GoogleCloudPlatform/gemini-cli)（Copyright 2025 Google LLC）を基に改変しています。

---

<a id="english"></a>

## English

**gmn-gui** is an unofficial GUI client for [Gemini CLI](https://github.com/GoogleCloudPlatform/gemini-cli). It provides AI-powered coding assistance as a native desktop app using Google's Gemini Code Assist API.
<img width="1046" height="751" alt="Screenshot 2026-02-06 at 3 00 31 PM" src="https://github.com/user-attachments/assets/d9980247-0c06-417c-af22-49cc65731428" />

### Features

- **Project Launcher** — Open recent projects with one click, restore past sessions
- **AI Chat** — Real-time streaming conversations with Gemini models
- **15 Built-in Tools** — File operations, shell execution, web search, grep, and more
- **Plan Mode** — Restrict AI to read-only tools for safe code exploration
- **MCP Server Support** — Extend tools via Model Context Protocol
- **Session Management** — Auto-save and restore conversations, organized by project
- **Multilingual** — Japanese / English UI with automatic system language detection
- **Theme Customization** — 8 accent colors, adjustable font size
- **API Usage Display** — Check remaining quota per model in real time
- **Markdown Rendering** — Code blocks with syntax highlighting

### Requirements

- macOS 12+ (Apple Silicon)
- [Gemini Code Assist](https://codeassist.google/) account (free with a Google account)

### Installation

#### Using a release build

1. Download `gmn-gui-darwin-arm64.zip` from [Releases](https://github.com/tomohiro-owada/gmn-gui/releases)
2. Unzip and move `gmn-gui.app` to `/Applications`

#### macOS Gatekeeper warning

gmn-gui is an unsigned app. On first launch, macOS will show: *"gmn-gui can't be opened because it is from an unidentified developer."*

**Workaround:**

```bash
# Run once in Terminal
xattr -cr /Applications/gmn-gui.app
```

Or:

1. In Finder, **right-click** (Control+click) `gmn-gui.app` → select "Open"
2. Click "Open" in the warning dialog

> Subsequent launches will work normally.

#### Building from source

```bash
# Prerequisites
# - Go 1.23+
# - Node.js 18+
# - Wails CLI v2

go install github.com/wailsapp/wails/v2/cmd/wails@latest

git clone https://github.com/tomohiro-owada/gmn-gui.git
cd gmn-gui
wails build
```

The built app will be at `build/bin/gmn-gui.app`.

### Usage

#### 1. First-time login

1. Launch the app — the **Launcher** window appears
2. Click the ⚙ icon in the top-right corner to open Settings
3. Click "Sign in with Google" and authenticate in your browser
4. Your project ID will be automatically retrieved after authentication

#### 2. Open a project

1. Click "Open Directory" to select a project folder
2. A chat window opens in a new window
3. Next time, you can open it with one click from the Launcher list

#### 3. Chat with AI

- Type a message and press Enter to send (Shift+Enter for newline)
- The AI automatically uses tools for file I/O, shell commands, web search, etc.
- Switch models using the selector in the top-right corner

#### 4. Plan Mode

- Toggle Plan Mode with the "Plan" button in the chat header
- When ON, the AI can only use read-only tools (file reading, grep, directory listing, etc.)
- Useful for safely investigating and analyzing code

#### 5. MCP Servers

- Manage connected servers from the MCP tab in the sidebar
- Configure servers in `~/.gemini/settings.json`:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "npx",
      "args": ["-y", "@example/mcp-server"]
    }
  }
}
```

#### 6. Settings

| Setting | Description |
|---------|-------------|
| Accent Color | Choose from 8 UI theme colors |
| Language | Japanese / English |
| Default Model | Model to use for new chats |
| Usage | Remaining quota and reset time per model |

#### 7. Slash commands

| Command | Description |
|---------|-------------|
| `/usage` or `/stats` | Show model usage and remaining quota |

### Supported Models

| Model | Description |
|-------|-------------|
| gemini-3-pro-preview | Highest accuracy (preview) |
| gemini-3-flash-preview | Fast & accurate (preview) |
| gemini-2.5-pro | High accuracy |
| gemini-2.5-flash | Standard (default) |
| gemini-2.5-flash-lite | Lightweight & fastest |

### Built-in Tools

| Tool | Description |
|------|-------------|
| `run_shell_command` | Execute shell commands |
| `read_file` | Read files (with optional line range) |
| `write_file` | Create or overwrite files |
| `replace` | Find and replace text in files |
| `read_many_files` | Read multiple files via glob pattern |
| `list_directory` | List directory contents |
| `glob` | Search files by pattern |
| `grep_search` | Regex search across files |
| `google_web_search` | Web search |
| `web_fetch` | Fetch URL content |
| `ask_user` | Ask user a question (choice/text) |
| `write_todos` | Save tasks to GEMINI.md |
| `save_memory` | Save notes to GEMINI.md |

### Configuration files

```
~/.gemini/
├── settings.json          # Global settings & MCP servers
├── oauth_creds.json       # Auth credentials (auto-generated)
├── gmn_state.json         # Cached state
└── gmn-gui/
    └── sessions/          # Chat sessions
```

### Tech Stack

| Layer | Technology |
|-------|-----------|
| Framework | [Wails v2](https://wails.io/) |
| Backend | Go 1.23 |
| Frontend | Vue 3 + TypeScript |
| Build | Vite 6 |
| Styling | Tailwind CSS 4 |
| UI Components | Radix Vue |
| State Management | Pinia |
| API | Gemini Code Assist (cloudcode-pa.googleapis.com) |

### License

Apache License 2.0

Portions of the code are modified from [Gemini CLI](https://github.com/GoogleCloudPlatform/gemini-cli) (Copyright 2025 Google LLC).
