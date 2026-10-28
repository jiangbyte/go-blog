import { resolve } from 'node:path'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import UnoCSS from 'unocss/vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'
import Icons from 'unplugin-icons/vite'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, __dirname, '')
  const port = Number(env.VITE_PORT || 5173)
  // Cookie 会话需同源；开发时由 Vite 反代到后端，避免跨域丢 Authorization。
  const proxyTarget = env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8000'
  // 与 VITE_BASE_URL / 路由 history base 一致；Docker 生产为 /admin/
  const base = env.VITE_BASE_URL || '/'

  return {
    base,
    server: {
      host: '0.0.0.0',
      port,
      proxy: {
        // mypage FastAPI：无 /api 前缀，按路径反代
        '/auth': { target: proxyTarget, changeOrigin: true },
        '/admin': { target: proxyTarget, changeOrigin: true },
        '/portal': { target: proxyTarget, changeOrigin: true },
        '/health': { target: proxyTarget, changeOrigin: true },
        '/uploads': { target: proxyTarget, changeOrigin: true },
      },
    },
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },
    optimizeDeps: {
      include: ['node-forge'],
    },
    build: {
      commonjsOptions: {
        transformMixedEsModules: true,
      },
    },
    plugins: [
      vue(),
      vueJsx(),
      UnoCSS(),
      Icons(),
      AutoImport({
        dts: 'src/typing/auto-imports.d.ts',
        imports: ['vue', 'pinia', 'vue-router', '@vueuse/core'],
      }),
      Components({
        dts: 'src/typing/components.d.ts',
        resolvers: [NaiveUiResolver()],
      }),
    ],
  }
})
