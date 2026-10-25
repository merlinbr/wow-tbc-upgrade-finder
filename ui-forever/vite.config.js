import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

export default defineConfig({
  base: './',
  plugins: [svelte()],
  build: {
    outDir: '../cmd/forever-armory/webui',
    emptyOutDir: true,
  },
});
