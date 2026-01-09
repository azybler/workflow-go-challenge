import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: true, // This enables listening on all addresses
    strictPort: true, // This ensures Vite fails if port 3000 is not available
    proxy: {
      '/api': {
        target: 'http://api:8080', // Route to the API container
        changeOrigin: true,
        secure: false,
      },
    },
  },
});
