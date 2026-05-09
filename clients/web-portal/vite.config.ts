import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      'guestPortal': path.resolve(__dirname, 'src/microfrontends/guestPortal'),
      'staffConsole': path.resolve(__dirname, 'src/microfrontends/staffConsole'),
      'executiveDashboard': path.resolve(__dirname, 'src/microfrontends/executiveDashboard'),
      'franchisePortal': path.resolve(__dirname, 'src/microfrontends/franchisePortal'),
      'mspPortal': path.resolve(__dirname, 'src/microfrontends/mspPortal'),
      'iptvManager': path.resolve(__dirname, 'src/microfrontends/iptvManager'),
      'iotControlPanel': path.resolve(__dirname, 'src/microfrontends/iotControlPanel'),
      'aiAnalytics': path.resolve(__dirname, 'src/microfrontends/aiAnalytics'),
    },
  },
  server: {
    port: 3000,
    host: true,
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})
