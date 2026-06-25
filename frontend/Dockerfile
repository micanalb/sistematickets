# ── Etapa 1: Compilación ─────────────────────────────────────────────
FROM node:20-alpine AS build

WORKDIR /app

# Copiamos primero package.json para aprovechar el cache de Docker:
# si las dependencias no cambian, no se vuelve a correr npm install.
COPY package*.json ./
RUN npm install

COPY . .

# Genera la build de producción optimizada en la carpeta dist/
RUN npm run build

# ── Etapa 2: Imagen final ────────────────────────────────────────────
# Reutilizamos la misma imagen de Node para servir la build con
# "vite preview" — más simple de explicar que configurar Nginx,
# a costa de no ser la forma más liviana para producción real.
FROM node:20-alpine

WORKDIR /app

# Solo necesitamos la build ya generada y los archivos de configuración
# mínimos para que "vite preview" sepa cómo servirla.
COPY --from=build /app/dist ./dist
COPY --from=build /app/package*.json ./
COPY --from=build /app/vite.config.js ./

# Instalamos solo las dependencias necesarias para correr vite preview
# (no hace falta volver a instalar TODAS las devDependencies de nuevo
# si vite ya quedó en node_modules de la etapa de build... pero como
# no copiamos node_modules, instalamos liviano acá)
RUN npm install

EXPOSE 5173

# --host 0.0.0.0 es necesario para que el servidor sea accesible desde
# fuera del contenedor (por defecto solo escucha en localhost interno)
CMD ["npx", "vite", "preview", "--host", "0.0.0.0", "--port", "5173"]