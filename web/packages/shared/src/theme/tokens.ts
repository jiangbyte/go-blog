/** Portal 主题：主色 #1677FF */
export const mypageSeedToken = {
  colorPrimary: '#1677FF',
  borderRadius: 6,
  colorBgLayout: '#f0f4fa',
  fontFamily: "'Source Sans 3', 'Noto Sans SC', ui-sans-serif, system-ui, sans-serif",
} as const

export const mypageTheme = {
  token: { ...mypageSeedToken },
  components: {
    Layout: {
      headerBg: '#ffffff',
      siderBg: '#0f1c2e',
    },
    Menu: {
      darkItemBg: '#0f1c2e',
      darkSubMenuItemBg: '#0b1522',
    },
  },
}
