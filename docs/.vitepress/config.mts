import { execFileSync } from 'node:child_process'
import { cpSync, mkdirSync, readFileSync, rmSync } from 'node:fs'
import { delimiter, dirname, join } from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'
import type { Plugin } from 'vite'

const vitepressDir = dirname(fileURLToPath(import.meta.url))
const docsDir = join(vitepressDir, '..')
const repoDir = join(docsDir, '..')
const playgroundDir = join(repoDir, 'playground')
const embeddedPlaygroundDir = join(docsDir, 'public', 'playground-app')
const binaryPath = [
  join(playgroundDir, 'node_modules', '.bin'),
  join(repoDir, 'node_modules', '.bin'),
  process.env.PATH ?? ''
].join(delimiter)
let embeddedPlaygroundBuilt = false

const runeGrammar = JSON.parse(
  readFileSync(
    new URL('../../vscode-rune/syntaxes/rune.tmLanguage.json', import.meta.url),
    'utf-8'
  )
)

const runeLanguage = {
  ...runeGrammar,
  name: 'rune',
  aliases: ['rn']
}

function buildEmbeddedPlayground() {
  execFileSync('pnpm', ['run', 'build'], {
    cwd: playgroundDir,
    env: {
      ...process.env,
      PATH: binaryPath,
      PLAYGROUND_BASE: '/playground-app/'
    },
    stdio: 'inherit'
  })
  rmSync(embeddedPlaygroundDir, { recursive: true, force: true })
  mkdirSync(dirname(embeddedPlaygroundDir), { recursive: true })
  cpSync(join(playgroundDir, 'dist'), embeddedPlaygroundDir, { recursive: true })
  embeddedPlaygroundBuilt = true
}

function ensureEmbeddedPlaygroundBuilt(force = false) {
  if (embeddedPlaygroundBuilt && !force) {
    return
  }
  buildEmbeddedPlayground()
}

function embeddedPlaygroundPlugin(): Plugin {
  let command: 'build' | 'serve' = 'build'
  const watched = [
    join(playgroundDir, 'src'),
    join(playgroundDir, 'index.html'),
    join(playgroundDir, 'vite.config.ts'),
    join(playgroundDir, 'package.json'),
    join(repoDir, 'cmd', 'rune-wasm'),
    join(repoDir, 'core'),
    join(repoDir, 'examples'),
    join(repoDir, 'internal')
  ]
  return {
    name: 'rune-docs-playground',
    configResolved(config) {
      command = config.command
    },
    buildStart() {
      if (command === 'build') {
        ensureEmbeddedPlaygroundBuilt()
      }
    },
    configureServer(server) {
      ensureEmbeddedPlaygroundBuilt()
      server.watcher.add(watched)
      server.watcher.on('change', (file) => {
        if (watched.some((path) => file.startsWith(path))) {
          ensureEmbeddedPlaygroundBuilt(true)
          server.ws.send({ type: 'full-reload' })
        }
      })
    }
  }
}

const englishSidebar = [
  {
    text: 'Guide',
    items: [
      { text: 'Overview', link: '/' },
      { text: 'Getting Started', link: '/guide/getting-started' }
    ]
  },
  {
    text: 'Language',
    items: [
      { text: 'Fundamentals', link: '/language/fundamentals' },
      { text: 'Core Library', link: '/language/core-library' }
    ]
  },
  {
    text: 'Tools',
    items: [
      { text: 'CLI and Editor', link: '/tools/cli-and-editor' },
      { text: 'Playground', link: '/playground' }
    ]
  }
]

const chineseSidebar = [
  {
    text: '指南',
    items: [
      { text: '概览', link: '/zh/' },
      { text: '快速开始', link: '/zh/guide/getting-started' }
    ]
  },
  {
    text: '语言',
    items: [
      { text: '基础语法', link: '/zh/language/fundamentals' },
      { text: '核心库', link: '/zh/language/core-library' }
    ]
  },
  {
    text: '工具',
    items: [
      { text: 'CLI 与编辑器', link: '/zh/tools/cli-and-editor' },
      { text: 'Playground', link: '/zh/playground' }
    ]
  }
]

export default {
  title: 'Rune',
  description: 'Documentation for the Rune programming language',
  cleanUrls: true,
  lastUpdated: false,
  head: [['link', { rel: 'icon', type: 'image/svg+xml', href: '/rune-icon.svg' }]],
  markdown: {
    theme: {
      light: 'github-light',
      dark: 'github-dark'
    },
    shikiSetup: async (shiki) => {
      await shiki.loadLanguage(runeLanguage)
    },
    // Keep compatibility with VitePress 1.x while moving to 2.x alpha.
    languages: [runeLanguage]
  },
  vite: {
    plugins: [embeddedPlaygroundPlugin()]
  },
  themeConfig: {
    logo: '/rune-icon.svg',
    search: { provider: 'local' },
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Language', link: '/language/fundamentals' },
      { text: 'Core Library', link: '/language/core-library' },
      { text: 'Playground', link: '/playground' }
    ],
    socialLinks: [{ icon: 'github', link: 'https://github.com/oboard/rune-lang' }],
    sidebar: englishSidebar
  },
  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      title: 'Rune',
      description: 'Documentation for the Rune programming language',
      themeConfig: {
        logo: '/rune-icon.svg',
        nav: [
          { text: 'Guide', link: '/guide/getting-started' },
          { text: 'Language', link: '/language/fundamentals' },
          { text: 'Core Library', link: '/language/core-library' },
          { text: 'Playground', link: '/playground' }
        ],
        socialLinks: [{ icon: 'github', link: 'https://github.com/oboard/rune-lang' }],
        sidebar: englishSidebar
      }
    },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      title: 'Rune',
      description: 'Rune 编程语言文档',
      themeConfig: {
        logo: '/rune-icon.svg',
        nav: [
          { text: '指南', link: '/zh/guide/getting-started' },
          { text: '语言', link: '/zh/language/fundamentals' },
          { text: '核心库', link: '/zh/language/core-library' },
          { text: 'Playground', link: '/zh/playground' }
        ],
        socialLinks: [{ icon: 'github', link: 'https://github.com/oboard/rune-lang' }],
        sidebar: chineseSidebar,
        outline: { label: '本页目录' },
        docFooter: { prev: '上一页', next: '下一页' },
        darkModeSwitchLabel: '外观',
        lightModeSwitchTitle: '切换到浅色模式',
        darkModeSwitchTitle: '切换到深色模式',
        sidebarMenuLabel: '菜单',
        returnToTopLabel: '回到顶部',
        langMenuLabel: '切换语言'
      }
    }
  }
}
