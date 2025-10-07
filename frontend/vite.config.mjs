import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  build: {
    rollupOptions: {
      external: ['src/test-connections.js']
    }
  },
  plugins: [react()],
  server: {
    host: process.env.FRONTEND_HOST,
    port: parseInt(process.env.FRONTEND_PORT),
    open: true,
  }
});
