import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    // Proxy para evitar CORS durante el desarrollo local (npm run dev).
    // /api reenvía las llamadas a la API, /uploads reenvía las imágenes
    // de eventos servidas como contenido estático por el backend.
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  },
  preview: {
    port: 5173,
    host: '0.0.0.0',
    // Mismo proxy que en "server", pero usado cuando se corre con
    // "vite preview" (lo que hace nuestro Dockerfile.frontend).
    // VITE_API_TARGET permite que docker-compose.yml indique el nombre
    // del servicio backend ("http://backend:8080") en vez de localhost,
    // ya que dentro de la red de contenedores no existe "localhost:8080".
    proxy: {
      '/api': {
        target: process.env.VITE_API_TARGET || 'http://localhost:8080',
        changeOrigin: true,
      },
      '/uploads': {
        target: process.env.VITE_API_TARGET || 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})