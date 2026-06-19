package clients

import (
	"fmt"
	"log"
	"os"
	"time"

	"ticketsya/domain"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// instanciaDB es el singleton de conexión a la base de datos, guarda la unica conexion a la BD
// empieza en nil(vacia)
var instanciaDB *gorm.DB

// ObtenerDB retorna la instancia singleton de la base de datos.
// Panics si la base de datos no fue inicializada previamente.
// Si alguien intenta usarla antes de inicializarla, el programa se detiene con log.Fatal.
// Esta función no se usa mucho porque pasamos la conexión directamente a los DAOs.
func ObtenerDB() *gorm.DB {
	if instanciaDB == nil {
		log.Fatal("Base de datos no inicializada. Llame a InicializarDB() primero.")
	}
	return instanciaDB
}

// InicializarDB establece la conexión con MySQL usando las variables de entorno
// y ejecuta la migración automática de las entidades del dominio con GORM.
func InicializarDB() (*gorm.DB, error) {
	// Lectura de variables de entorno para la cadena de conexión desde el .env
	host := obtenerEnv("DB_HOST", "localhost")
	puerto := obtenerEnv("DB_PORT", "3306")
	usuario := obtenerEnv("DB_USER", "root")
	password := obtenerEnv("DB_PASSWORD", "root")
	nombre := obtenerEnv("DB_NAME", "sistematickets")

	// Construcción del DSN (Data Source Name) para MySQL, cadena de conexion
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		usuario, password, host, puerto, nombre,
	)

	// Configuración del logger de GORM según entorno
	//En modo development GORM muestra todas las queries SQL en la terminal (eso es lo que ves cuando corre).
	//En producción las silencia con logger.Silent para no llenar los logs.
	nivelLog := logger.Silent
	if os.Getenv("APP_ENV") == "development" {
		nivelLog = logger.Info
	}
	//Intenta conectarse a MySQL con el DSN. Si falla (contraseña incorrecta, MySQL no corre, etc)
	//retorna el error y el programa no arranca
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(nivelLog),
	})
	if err != nil {
		return nil, fmt.Errorf("error al conectar con MySQL: %w", err)
	}

	// Configuración del pool de conexiones
	//Configura cuantas conexiones simultaneas puede manejar
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)           //maximo 10 conexiones en espera sin usar
	sqlDB.SetMaxOpenConns(100)          //maximo 100 conexiones abiertas al mismo tiempo
	sqlDB.SetConnMaxLifetime(time.Hour) //cada conexion se recicla despues de 1 hora
	//Evita que el sistema colapse si hay muchos usuarios al mismo tiempo

	// Auto-migración: GORM crea/actualiza las tablas según los structs del dominio
	// Orden: primero entidades sin dependencias (Usuario y Evento), luego las que tienen FK (Entradas)
	err = db.AutoMigrate(
		&domain.Usuario{},
		&domain.Evento{},
		&domain.Entrada{},
	)
	if err != nil {
		return nil, fmt.Errorf("error en auto-migración de GORM: %w", err)
	}
	//guarda la conexion en el singleton
	instanciaDB = db
	log.Println("✅ Base de datos conectada y migrada correctamente")
	return db, nil
}

// obtenerEnv retorna el valor de una variable de entorno o un valor por defecto (si no existe o esta vacía)
func obtenerEnv(clave, valorDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return valorDefecto
}
