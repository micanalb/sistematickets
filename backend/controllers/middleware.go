package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"ticketsya/domain"
	"ticketsya/utils"
)

const (
	// ClaveUsuarioID es la clave para acceder al ID del usuario en el contexto de Gin
	ClaveUsuarioID = "usuario_id"
	// ClaveUsuarioRol es la clave para acceder al rol del usuario en el contexto de Gin
	ClaveUsuarioRol = "usuario_rol"
	// ClaveUsuarioEmail es la clave para acceder al email en el contexto de Gin
	ClaveUsuarioEmail = "usuario_email"
)

// MiddlewareAutenticacion valida el token JWT en el header Authorization.
// Si el token es válido, inyecta los datos del usuario en el contexto de Gin.
// Se usa en rutas que requieren autenticación.
func MiddlewareAutenticacion() gin.HandlerFunc {
	return func(c *gin.Context) {
		// El token debe venir en el header: Authorization: Bearer <token>
		headerAuth := c.GetHeader("Authorization")
		if headerAuth == "" {
			utils.ResponderNoAutorizado(c, "se requiere token de autenticación")
			c.Abort()
			return
		}

		// Extraer el token del formato "Bearer <token>"
		partes := strings.SplitN(headerAuth, " ", 2)
		if len(partes) != 2 || strings.ToLower(partes[0]) != "bearer" {
			utils.ResponderNoAutorizado(c, "formato de token inválido. Use: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := partes[1]

		// Validar y parsear el token
		claims, err := utils.ValidarTokenJWT(tokenString)
		if err != nil {
			utils.ResponderNoAutorizado(c, "token inválido o expirado")
			c.Abort()
			return
		}

		// Inyectar datos del usuario en el contexto para uso en los controladores
		c.Set(ClaveUsuarioID, claims.UsuarioID)
		c.Set(ClaveUsuarioRol, claims.Rol)
		c.Set(ClaveUsuarioEmail, claims.Email)

		c.Next()
	}
}

// MiddlewareRolAdmin verifica que el usuario autenticado sea administrador.
// Debe usarse DESPUÉS de MiddlewareAutenticacion en la cadena de middlewares.
func MiddlewareRolAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		rol, existe := c.Get(ClaveUsuarioRol)
		if !existe {
			utils.ResponderNoAutorizado(c, "usuario no autenticado")
			c.Abort()
			return
		}

		if rol != domain.RolAdministrador {
			utils.ResponderForbidden(c, "acceso denegado: se requiere rol de administrador")
			c.Abort()
			return
		}

		c.Next()
	}
}

// obtenerUsuarioIDDelContexto es un helper para extraer el ID del usuario del contexto Gin
func obtenerUsuarioIDDelContexto(c *gin.Context) (uint, bool) {
	valor, existe := c.Get(ClaveUsuarioID)
	if !existe {
		return 0, false
	}
	id, ok := valor.(uint)
	return id, ok
}
