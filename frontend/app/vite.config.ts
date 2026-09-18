import path from 'node:path'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Keep local browser requests same-origin while AgentGate runs on :8090.
      '/api': 'http://127.0.0.1:8090',
    },
  },
  resolve: {
    alias: {
      // '@contract' points at the existing, already-tested G1-G4 contract layer
      // (models, api client, state store, view renderers). The UI imports from
      // here rather than redefining any of these types or logic.
      '@contract': path.resolve(import.meta.dirname, '../src'),
    },
  },
})
