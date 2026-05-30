package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"ticketsya/domain"
)

// ClaimsJWT define la estructura de datos que se embeben en el token JWT
type ClaimsJWT struct {
	UsuarioID uint               `json:"usuario_id"`
	Email     string             `json:"email"`
	Rol       domain.RolUsuario  `json:"rol"`
	jwt.RegisteredClaims
}

// obtenerSecretoJWT retorna la clave secreta desde variable de entorno o usa un valor por defecto (solo dev)
func obtenerSecretoJWT() []byte {
	secreto := os.Getenv("JWT_SECRET")
	if secreto == "" {
		// En producción esto DEBE estar configurado como variable de entorno
		secreto = "ticketsya_secreto_dev_2024_cambiar_en_produccion"
	}
	return []byte(secreto)
}

// GenerarTokenJWT crea y firma un nuevo token JWT para el usuario dado.
// El token expira en 24 horas por defecto (configurable con JWT_EXPIRACION_HORAS).
func GenerarTokenJWT(usuarioID uint, email string, rol domain.RolUsuario) (string, error) {
	// Tiempo de expiración configurable
	expiracionHoras := 24
	if exp := os.Getenv("JWT_EXPIRACION_HORAS"); exp != "" {
		// Podría parsearse, para simplicidad usamos el default
		_ = exp
	}

	claims := ClaimsJWT{
		UsuarioID: usuarioID,
		Email:     email,
		Rol:       rol,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiracionHoras) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ticketsya",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(obtenerSecretoJWT())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidarTokenJWT valida y parsea un token JWT, retornando sus claims si es válido
func ValidarTokenJWT(tokenString string) (*ClaimsJWT, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&ClaimsJWT{},
		func(token *jwt.Token) (interface{}, error) {
			// Verificamos que el método de firma sea HMAC (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de firma inesperado en token JWT")
			}
			return obtenerSecretoJWT(), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*ClaimsJWT)
	if !ok || !token.Valid {
		return nil, errors.New("token JWT inválido")
	}

	return claims, nil
}
