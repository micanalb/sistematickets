package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"ticketsya/clients"
	"ticketsya/controllers"
	"ticketsya/dao"
	"ticketsya/services"
)

func main() {
	// Cargar variables de entorno desde .env
    	godotenv.Load()
	// Inicializar la conexión a la base de datos
	db, err := clients.InicializarDB()
	if err != nil {
		log.Fatalf("❌ Error al inicializar la base de datos: %v", err)
	}

	// ── Inicializar DAOs (capa de acceso a datos) ──────────────────────────────
	usuarioDAO := dao.NuevoUsuarioDAO(db)
	eventoDAO := dao.NuevoEventoDAO(db)
	entradaDAO := dao.NuevoEntradaDAO(db)

	// ── Inicializar Servicios (lógica de negocio) ──────────────────────────────
	authService := services.NuevoAuthService(usuarioDAO)
	eventoService := services.NuevoEventoService(eventoDAO)
	entradaService := services.NuevoEntradaService(entradaDAO, eventoDAO, usuarioDAO)

	// ── Inicializar Controladores ──────────────────────────────────────────────
	authController := controllers.NuevoAuthController(authService)
	eventoController := controllers.NuevoEventoController(eventoService)
	entradaController := controllers.NuevoEntradaController(entradaService)

	// ── Configurar Gin según el entorno ────────────────────────────────────────
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// ── Middlewares globales ───────────────────────────────────────────────────
	// CORS: permitir peticiones desde el frontend
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
		if c.GetHeader("Access-Control-Allow-Origin") == "" {
			c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// ── Definición de rutas ────────────────────────────────────────────────────
	// Todas las rutas están bajo el prefijo /api/v1
	api := router.Group("/api/v1")
	{
		authController.RegistrarRutas(api)
		eventoController.RegistrarRutas(api)
		entradaController.RegistrarRutas(api)
	}

	// Ruta de health check para verificar que el servidor está corriendo
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"estado":  "ok",
			"version": "1.0.0",
			"app":     "TicketsYa",
		})
	})

	// ── Arrancar el servidor ───────────────────────────────────────────────────
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}

	log.Printf("🚀 Servidor TicketsYa corriendo en http://localhost:%s", puerto)
	log.Printf("📡 Ambiente: %s", os.Getenv("APP_ENV"))

	if err := router.Run(":" + puerto); err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}
