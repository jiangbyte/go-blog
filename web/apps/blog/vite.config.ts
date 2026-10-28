import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, __dirname, '')
  const port = Number(env.VITE_PORT || 5173)
  const proxyTarget = env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8000'
  const base = env.VITE_BASE_URL || '/'

  return {
    base,
    plugins: [vue(), tailwindcss()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      host: '0.0.0.0',
      port,
      proxy: {
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
          rewrite: (p) => p.replace(/^\/api/, ''),
        },
        '/uploads': { target: proxyTarget, changeOrigin: true },
      },
    },
  }
})
