package main

import (
	"log"
	"os"

	"ticketsya/clients"
	"ticketsya/controllers"
	"ticketsya/dao"
	"ticketsya/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno desde .env
	//Lee el archivo .env y carga esas variables como variables de entorno del proceso. En Go esto es necesario hacerlo explícitamente
	godotenv.Load()
	// Inicializar la conexión a la base de datos
	//Se conecta a MySQL. Si falla, log.Fatalf corta la ejecución del programa inmediatamente (loguea el error y hace os.Exit(1)).
	db, err := clients.InicializarDB()
	if err != nil {
		log.Fatalf("❌ Error al inicializar la base de datos: %v", err)
	}

	//DAO  →  Service  →  Controller

	// ── Inicializar DAOs (capa de acceso a datos) ──────────────────────────────
	usuarioDAO := dao.NuevoUsuarioDAO(db)
	eventoDAO := dao.NuevoEventoDAO(db)
	entradaDAO := dao.NuevoEntradaDAO(db)

	//la capa que habla directo con la base de datos (queries SQL vía GORM).
	// Recibe db porque necesita ejecutar consultas.

	// ── Inicializar Servicios (lógica de negocio) ──────────────────────────────
	authService := services.NuevoAuthService(usuarioDAO)
	eventoService := services.NuevoEventoService(eventoDAO)
	entradaService := services.NuevoEntradaService(entradaDAO, eventoDAO, usuarioDAO)

	//La lógica de negocio. Por ejemplo, entradaService recibe tres DAOs (entradaDAO, eventoDAO, usuarioDAO)
	// porque necesita validar cosas como "¿existe el evento?" o "¿existe el usuario?" antes de crear una entrada.

	// ── Inicializar Controladores ──────────────────────────────────────────────
	authController := controllers.NuevoAuthController(authService)
	eventoController := controllers.NuevoEventoController(eventoService)
	entradaController := controllers.NuevoEntradaController(entradaService)
	//la capa que recibe las peticiones HTTP y llama al service correspondiente.

	//cada capa recibe la capa anterior como parámetro en su constructor, en vez de crearla ella misma.
	// Facilita testear cada capa por separado

	// ── Configurar Gin según el entorno ────────────────────────────────────────
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// ── Middlewares globales ───────────────────────────────────────────────────
	// CORS: permitir peticiones desde el frontend
	router.Use(func(c *gin.Context) {
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}
		c.Header("Access-Control-Allow-Origin", frontendURL)
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
	//Toma el puerto de PORT, si no está seteado usa 8080 por default, y arranca con router.Run.
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
