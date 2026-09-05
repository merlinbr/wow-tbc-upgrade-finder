import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

const outDir = fileURLToPath(new URL('../cmd/wowsimcli/cmd/upgrade_ui/', import.meta.url));

export default defineConfig({
  base: './',
  server: { proxy: { '/api': 'http://127.0.0.1:43123' } },
  plugins: [svelte()],
  build: { outDir, emptyOutDir: true, assetsDir: 'assets' },
});
