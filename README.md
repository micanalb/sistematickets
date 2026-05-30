# 🎫 SistemaTickets

> Plataforma de venta de entradas para eventos — Primera entrega (Regularidad).

---

## Tecnologías

**Backend:** Go 1.21 · Gin · GORM · MySQL · JWT · bcrypt · testify  
**Frontend:** React 18 · Vite · React Router v6 · Axios

---

## Requisitos previos

- [Go](https://go.dev/dl/) 1.21+
- [Node.js](https://nodejs.org/) 20+
- [MySQL](https://dev.mysql.com/downloads/) 8.0+

---

## Instalación y uso

### 1. Base de datos

```sql
CREATE DATABASE sistematickets CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. Backend

```bash
cd backend

# Configurar variables de entorno
cp .env.example .env
# Editar .env con tus datos de MySQL

# Descargar dependencias
go mod download

# Ejecutar servidor (crea las tablas automáticamente)
go run main.go

# Servidor en: http://localhost:8080
```

### 3. Frontend

```bash
cd frontend

# Instalar dependencias
npm install

# Iniciar servidor de desarrollo
npm run dev

# Frontend en: http://localhost:5173
```

---

## Ejecutar los tests

```bash
cd backend

# Correr todos los tests
go test ./...

# Con detalle
go test -v ./...

# Ver cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## Endpoints de la API

Base URL: `http://localhost:8080/api/v1`

| Método | Endpoint | Descripción | Auth |
|---|---|---|---|
| POST | `/auth/registro` | Registrar usuario | No |
| POST | `/auth/login` | Iniciar sesión | No |
| GET | `/eventos` | Listar eventos (filtros opcionales) | No |
| GET | `/eventos/:id` | Detalle de un evento | No |
| POST | `/entradas/comprar` | Comprar entrada | JWT |
| GET | `/entradas/mis-entradas` | Mis entradas | JWT |
| PUT | `/entradas/:id/cancelar` | Cancelar entrada | JWT |
| PUT | `/entradas/:id/transferir` | Transferir entrada | JWT |

**Filtros disponibles en `/eventos`:**
```
?busqueda=rock
?categoria=musica|deporte|cultura|teatro_cine|conferencia|otro
?solo_disponibles=true
```

**Header de autenticación:**
```
Authorization: Bearer <token_jwt>
```

---

## Estructura del proyecto

```
sistematickets/
├── backend/
│   ├── clients/       # Conexión MySQL (GORM singleton)
│   ├── controllers/   # Handlers HTTP + Middlewares
│   ├── dao/           # Data Access Objects
│   ├── domain/        # Entidades y DTOs
│   ├── services/      # Lógica de negocio
│   ├── utils/         # JWT, bcrypt, respuestas, QR
│   └── main.go
└── frontend/
    └── src/
        ├── components/ # Navbar
        ├── context/    # AuthContext
        ├── pages/      # Inicio, Detalle, Login, Registro, MisEntradas
        └── services/   # api.js (Axios)
```

---

## Decisiones de diseño

**1. bcrypt en lugar de MD5/SHA256**  
Bcrypt incluye salt automático y es computacionalmente costoso por diseño, lo que lo hace mucho más seguro para almacenar contraseñas que MD5 o SHA256.

**2. Soft-delete para eventos**  
Los eventos eliminados conservan `deleted_at` pero no se borran físicamente. Esto preserva el historial de entradas compradas: si un evento se da de baja, los usuarios siguen viendo sus tickets en "Mis Entradas".

**3. Transferencia crea nueva entrada**  
Al transferir, la entrada original queda como "transferida" y se genera una nueva con QR distinto para el destinatario. Garantiza trazabilidad completa de la cadena de titularidad.

**4. Interfaces en DAOs y Servicios**  
Todas las capas se definen como interfaces, permitiendo mockear dependencias en tests sin necesitar una base de datos real.
