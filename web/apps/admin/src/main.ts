/** Author: go-blog */

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import {
  create,
  ProCard,
  ProDataTable,
  ProInput,
  ProModalForm,
  ProPassword,
  ProRadioGroup,
  ProSearchForm,
  ProSelect,
  ProTextarea,
} from 'pro-naive-ui'
import 'virtual:uno.css'
import './style.css'
import './plugins/iconify'
import App from './App.vue'
import { installRouter } from './router'

async function bootstrap() {
  const app = createApp(App)
  const pinia = createPinia()

  pinia.use(piniaPluginPersistedstate)
  app.use(pinia)
  app.use(
    create({
      components: [
        ProCard,
        ProSearchForm,
        ProDataTable,
        ProModalForm,
        ProInput,
        ProRadioGroup,
        ProPassword,
        ProSelect,
        ProTextarea,
      ],
    }),
  )
  await installRouter(app)
  // 启动时预拉站点品牌，供 Logo / 标题栏使用
  const { useSiteStore } = await import('./stores/site')
  void useSiteStore().ensureLoaded()
  app.mount('#app')
}

void bootstrap()
