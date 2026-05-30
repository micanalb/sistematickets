package controllers

import (
	"github.com/gin-gonic/gin"
	"ticketsya/domain"
	"ticketsya/services"
	"ticketsya/utils"
)

// AuthController maneja los endpoints de registro y login
type AuthController struct {
	authService services.AuthService
}

// NuevoAuthController crea una nueva instancia del controlador de auth
func NuevoAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// RegistrarRutas registra las rutas de autenticación en el router
func (ctrl *AuthController) RegistrarRutas(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/registro", ctrl.Registrar)
		auth.POST("/login", ctrl.Login)
	}
}

// Registrar godoc
// @Summary Registro de nuevo usuario
// @Description Crea una nueva cuenta de usuario en el sistema
// @Tags autenticación
// @Accept json
// @Produce json
// @Param body body domain.DTORegistro true "Datos del nuevo usuario"
// @Success 201 {object} domain.DTORespuestaLogin
// @Router /auth/registro [post]
func (ctrl *AuthController) Registrar(c *gin.Context) {
	var dto domain.DTORegistro

	// Binding y validación automática de los campos requeridos
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.ResponderBadRequest(c, "datos de registro inválidos: "+err.Error())
		return
	}

	respuesta, err := ctrl.authService.Registrar(dto)
	if err != nil {
		// Distinguir error de email duplicado vs error interno
		utils.ResponderConflicto(c, err.Error())
		return
	}

	utils.ResponderCreado(c, respuesta)
}

// Login godoc
// @Summary Login de usuario
// @Description Autentica al usuario y retorna un token JWT
// @Tags autenticación
// @Accept json
// @Produce json
// @Param body body domain.DTOLogin true "Credenciales"
// @Success 200 {object} domain.DTORespuestaLogin
// @Router /auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var dto domain.DTOLogin

	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.ResponderBadRequest(c, "datos de login inválidos: "+err.Error())
		return
	}

	respuesta, err := ctrl.authService.Login(dto)
	if err != nil {
		// Mensaje genérico para no revelar si el email existe
		utils.ResponderNoAutorizado(c, "credenciales inválidas")
		return
	}

	utils.ResponderExito(c, respuesta)
}
