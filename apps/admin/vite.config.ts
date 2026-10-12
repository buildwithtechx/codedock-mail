import tailwindcss from '@tailwindcss/vite';
import { tanstackRouter } from '@tanstack/router-plugin/vite';
import viteReact from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig({
  resolve: { tsconfigPaths: true },
  plugins: [tailwindcss(), tanstackRouter(), viteReact()],
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
