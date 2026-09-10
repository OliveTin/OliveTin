import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  resolve: {
    dedupe: ['vue', 'vue-router']
  },
  plugins: [vue()],
  build: {
    rolldownOptions: {
      onLog (level, log, defaultHandler) {
        if (log.code === 'INVALID_ANNOTATION') {
          return
        }
        defaultHandler(level, log)
      }
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:1337',
        changeOrigin: true,
        secure: false
      },
      '/theme.css': {
        target: 'http://localhost:1337',
        changeOrigin: true,
        secure: false
      },
      '/custom-webui': {
        target: 'http://localhost:1337',
        changeOrigin: true
      }
    }
  }
})
