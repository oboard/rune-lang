import { readFileSync } from 'node:fs'

const runeGrammar = JSON.parse(
  readFileSync(
    new URL('../../vscode-rune/syntaxes/rune.tmLanguage.json', import.meta.url),
    'utf-8'
  )
)

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
    items: [{ text: 'CLI and Editor', link: '/tools/cli-and-editor' }]
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
    items: [{ text: 'CLI 与编辑器', link: '/zh/tools/cli-and-editor' }]
  }
]

export default {
  title: 'Rune',
  description: 'Documentation for the Rune programming language',
  cleanUrls: true,
  lastUpdated: false,
  head: [['link', { rel: 'icon', type: 'image/svg+xml', href: '/rune-icon.svg' }]],
  markdown: {
    languages: [
      {
        name: 'rune',
        scopeName: 'source.rune',
        grammar: runeGrammar,
        aliases: ['rn']
      }
    ]
  },
  themeConfig: {
    logo: '/rune-icon.svg',
    search: { provider: 'local' },
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Language', link: '/language/fundamentals' },
      { text: 'Core Library', link: '/language/core-library' }
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
          { text: 'Core Library', link: '/language/core-library' }
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
          { text: '核心库', link: '/zh/language/core-library' }
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
