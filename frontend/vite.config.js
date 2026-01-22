import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      '/users': 'http://localhost:8080',
      '/posts': 'http://localhost:8080',
      '/auth': 'http://localhost:8080'
    }
  }
})
