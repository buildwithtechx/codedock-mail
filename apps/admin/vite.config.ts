import viteReact from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [viteReact()],
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
      },
    },
  },
  build: {
    outDir: './dist',
    emptyOutDir: true,
  },
});
