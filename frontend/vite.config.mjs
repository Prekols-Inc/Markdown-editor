import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import fs from 'fs';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  server: {
    host: process.env.FRONTEND_HOST,
    port: parseInt(process.env.FRONTEND_PORT),
    open: true,
    https: {
      key: fs.readFileSync(path.resolve(__dirname, 'tls/key.crt')),
      cert: fs.readFileSync(path.resolve(__dirname, 'tls/cert_frontend.crt')),
    }
  }
});
