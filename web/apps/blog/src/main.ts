import { createApp } from 'vue'
import App from './App.vue'
import { bootstrapApi } from './api'
import router from './router'
import './plugins/iconify'
import '@icon-park/vue-next/styles/index.css'
import 'katex/dist/katex.min.css'
import 'highlight.js/styles/github.css'
import './styles/theme.css'
import '@jiangbyte/markdown/styles/prose-base.css'
import '@jiangbyte/markdown/styles/prose.css'

await bootstrapApi()
createApp(App).use(router).mount('#app')
