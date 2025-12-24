import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import fs from 'fs';
import path from 'path';

export default defineConfig(() => {
  const certDir = process.env.CERT_DIR_PATH;
  const keyPath = certDir ? path.join(certDir, 'key.crt') : undefined;
  const certPath = certDir ? path.join(certDir, 'cert.crt') : undefined;

  const httpsConfig =
    keyPath &&
    certPath &&
    fs.existsSync(keyPath) &&
    fs.existsSync(certPath)
      ? {
          key: fs.readFileSync(keyPath),
          cert: fs.readFileSync(certPath),
        }
      : undefined;

  return {
    plugins: [react()],
    server: {
      host: process.env.FRONTEND_HOST,
      port: Number(process.env.FRONTEND_PORT),
      open: true,
      https: httpsConfig,
    },
  };
});
