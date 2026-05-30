package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ticketsya/controllers"
	"ticketsya/domain"
)

// ════════════════════════════════════════════════════════════════════
// Mock del AuthService para tests de controlador
// ════════════════════════════════════════════════════════════════════

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Registrar(dto domain.DTORegistro) (*domain.DTORespuestaLogin, error) {
	args := m.Called(dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DTORespuestaLogin), args.Error(1)
}

func (m *MockAuthService) Login(dto domain.DTOLogin) (*domain.DTORespuestaLogin, error) {
	args := m.Called(dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DTORespuestaLogin), args.Error(1)
}

// setupRouterAuth configura un router de test con el controlador de auth
func setupRouterAuth(svc *MockAuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NuevoAuthController(svc)
	api := router.Group("/api/v1")
	ctrl.RegistrarRutas(api)
	return router
}

// ════════════════════════════════════════════════════════════════════
// Tests HTTP de Registro
// ════════════════════════════════════════════════════════════════════

func TestRegistrar_HTTP_201_ExitoNuevoUsuario(t *testing.T) {
	mockSvc := new(MockAuthService)
	router := setupRouterAuth(mockSvc)

	respuestaEsperada := &domain.DTORespuestaLogin{
		Token: "token.jwt.mock",
		Usuario: domain.UsuarioPublico{
			ID:    1,
			Email: "nuevo@example.com",
			Rol:   domain.RolCliente,
		},
	}

	mockSvc.On("Registrar", mock.AnythingOfType("domain.DTORegistro")).Return(respuestaEsperada, nil)

	cuerpo, _ := json.Marshal(map[string]string{
		"nombre":    "Nuevo",
		"apellido":  "Usuario",
		"email":     "nuevo@example.com",
		"password":  "password123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/registro", bytes.NewBuffer(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var respuesta map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respuesta)
	assert.True(t, respuesta["exito"].(bool))
}

func TestRegistrar_HTTP_400_CuerpoInvalido(t *testing.T) {
	mockSvc := new(MockAuthService)
	router := setupRouterAuth(mockSvc)

	// JSON malformado
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/registro", bytes.NewBuffer([]byte("{malformed}")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegistrar_HTTP_409_EmailDuplicado(t *testing.T) {
	mockSvc := new(MockAuthService)
	router := setupRouterAuth(mockSvc)

	// El servicio retorna error de email duplicado
	mockSvc.On("Registrar", mock.AnythingOfType("domain.DTORegistro")).
		Return(nil, assert.AnError)

	cuerpo, _ := json.Marshal(map[string]string{
		"nombre":   "Test",
		"apellido": "User",
		"email":    "duplicado@example.com",
		"password": "password123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/registro", bytes.NewBuffer(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// ════════════════════════════════════════════════════════════════════
// Tests HTTP de Login
// ════════════════════════════════════════════════════════════════════

func TestLogin_HTTP_200_ExitoConCredencialesValidas(t *testing.T) {
	mockSvc := new(MockAuthService)
	router := setupRouterAuth(mockSvc)

	respuestaEsperada := &domain.DTORespuestaLogin{
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
		Usuario: domain.UsuarioPublico{
			ID:    1,
			Email: "usuario@example.com",
			Rol:   domain.RolCliente,
		},
	}

	mockSvc.On("Login", mock.AnythingOfType("domain.DTOLogin")).Return(respuestaEsperada, nil)

	cuerpo, _ := json.Marshal(map[string]string{
		"email":    "usuario@example.com",
		"password": "password123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respuesta map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respuesta)
	assert.True(t, respuesta["exito"].(bool))

	// Verificar que el token está en la respuesta
	datos := respuesta["datos"].(map[string]interface{})
	assert.NotEmpty(t, datos["token"])
}

func TestLogin_HTTP_401_CredencialesInvalidas(t *testing.T) {
	mockSvc := new(MockAuthService)
	router := setupRouterAuth(mockSvc)

	mockSvc.On("Login", mock.AnythingOfType("domain.DTOLogin")).
		Return(nil, assert.AnError)

	cuerpo, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "passwordMalo",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_HTTP_400_CamposRequeridosFaltantes(t *testing.T) {
	mockSvc := new(MockAuthService)
	router := setupRouterAuth(mockSvc)

	// Solo email, sin password
	cuerpo, _ := json.Marshal(map[string]string{
		"email": "test@example.com",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
