package clients

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"ticketsya/domain"
)

// instanciaDB es el singleton de conexión a la base de datos
var instanciaDB *gorm.DB

// ObtenerDB retorna la instancia singleton de la base de datos.
// Panics si la base de datos no fue inicializada previamente.
func ObtenerDB() *gorm.DB {
	if instanciaDB == nil {
		log.Fatal("Base de datos no inicializada. Llame a InicializarDB() primero.")
	}
	return instanciaDB
}

// InicializarDB establece la conexión con MySQL usando las variables de entorno
// y ejecuta la migración automática de las entidades del dominio con GORM.
func InicializarDB() (*gorm.DB, error) {
	// Lectura de variables de entorno para la cadena de conexión
	host := obtenerEnv("DB_HOST", "localhost")
	puerto := obtenerEnv("DB_PORT", "3306")
	usuario := obtenerEnv("DB_USER", "root")
	password := obtenerEnv("DB_PASSWORD", "root")
	nombre := obtenerEnv("DB_NAME", "ticketsya")

	// Construcción del DSN (Data Source Name) para MySQL
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		usuario, password, host, puerto, nombre,
	)

	// Configuración del logger de GORM según entorno
	nivelLog := logger.Silent
	if os.Getenv("APP_ENV") == "development" {
		nivelLog = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(nivelLog),
	})
	if err != nil {
		return nil, fmt.Errorf("error al conectar con MySQL: %w", err)
	}

	// Configuración del pool de conexiones
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migración: GORM crea/actualiza las tablas según los structs del dominio
	// Orden importante: primero entidades sin dependencias, luego las que tienen FK
	err = db.AutoMigrate(
		&domain.Usuario{},
		&domain.Evento{},
		&domain.Entrada{},
	)
	if err != nil {
		return nil, fmt.Errorf("error en auto-migración de GORM: %w", err)
	}

	instanciaDB = db
	log.Println("✅ Base de datos conectada y migrada correctamente")
	return db, nil
}

// obtenerEnv retorna el valor de una variable de entorno o un valor por defecto
func obtenerEnv(clave, valorDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return valorDefecto
}
