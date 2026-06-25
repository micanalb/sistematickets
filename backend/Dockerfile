# ── Etapa 1: Compilación ─────────────────────────────────────────────
# Usamos la imagen oficial de Go para compilar el binario. Esta etapa
# NO viaja a producción — solo se usa para construir el ejecutable.
FROM golang:1.21-alpine AS build

# Alpine no trae gcc por defecto; lo necesitamos porque alguna dependencia
# (drivers de BD, etc.) puede requerir CGO en tiempo de build.
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copiamos primero solo los archivos de dependencias. Docker cachea esta
# capa: si go.mod/go.sum no cambian, no vuelve a descargar todo en cada build.
COPY go.mod go.sum ./
RUN go mod download

# Ahora copiamos el resto del código fuente
COPY . .

# Compilamos un binario estático (CGO_ENABLED=0) para que la imagen final
# no necesite ninguna librería del sistema operativo.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/servidor main.go

# ── Etapa 2: Imagen final ────────────────────────────────────────────
# Alpine es una distribución mínima (~5MB) — la imagen final pesa muy
# poco porque no incluye el compilador de Go ni el código fuente, solo
# el binario ya compilado.
FROM alpine:latest

WORKDIR /app

# Copiamos únicamente el binario compilado desde la etapa anterior
COPY --from=build /app/servidor .

# Carpeta donde se guardan las imágenes subidas de eventos. Se crea acá
# para que exista incluso antes del primer request; en docker-compose.yml
# se monta como volumen para persistir entre reinicios del contenedor.
RUN mkdir -p /app/uploads/eventos

EXPOSE 8080

CMD ["./servidor"]