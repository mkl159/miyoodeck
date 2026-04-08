import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'


export default defineConfig({
  plugins: [svelte()],
  build: {
    outDir: '../package/App/WebDeck/www',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: undefined,
        // Single JS bundle for simplicity
        inlineDynamicImports: true,
      }
    }
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': { target: 'ws://localhost:8080', ws: true }
    }
  }
})
