package services_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ticketsya/domain"
	"ticketsya/services"
	"ticketsya/utils"
)

// ════════════════════════════════════════════════════════════════════
// Mock del UsuarioDAO para aislar los tests del servicio de auth
// ════════════════════════════════════════════════════════════════════

type MockUsuarioDAO struct {
	mock.Mock
}

func (m *MockUsuarioDAO) Crear(usuario *domain.Usuario) error {
	args := m.Called(usuario)
	return args.Error(0)
}

func (m *MockUsuarioDAO) BuscarPorID(id uint) (*domain.Usuario, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Usuario), args.Error(1)
}

func (m *MockUsuarioDAO) BuscarPorEmail(email string) (*domain.Usuario, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Usuario), args.Error(1)
}

func (m *MockUsuarioDAO) Actualizar(usuario *domain.Usuario) error {
	args := m.Called(usuario)
	return args.Error(0)
}

func (m *MockUsuarioDAO) ExisteEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

// ════════════════════════════════════════════════════════════════════
// Tests de Registro
// ════════════════════════════════════════════════════════════════════

func TestRegistrar_ExitosoConDatosValidos(t *testing.T) {
	mockDAO := new(MockUsuarioDAO)
	svc := services.NuevoAuthService(mockDAO)

	dto := domain.DTORegistro{
		Nombre:   "Juan",
		Apellido: "Pérez",
		Email:    "juan@example.com",
		Password: "password123",
	}

	// El email no existe aún
	mockDAO.On("ExisteEmail", dto.Email).Return(false, nil)
	// La creación es exitosa
	mockDAO.On("Crear", mock.AnythingOfType("*domain.Usuario")).Return(nil)

	respuesta, err := svc.Registrar(dto)

	assert.NoError(t, err)
	assert.NotNil(t, respuesta)
	assert.NotEmpty(t, respuesta.Token)
	assert.Equal(t, dto.Email, respuesta.Usuario.Email)
	assert.Equal(t, domain.RolCliente, respuesta.Usuario.Rol)
	mockDAO.AssertExpectations(t)
}

func TestRegistrar_ErrorEmailDuplicado(t *testing.T) {
	mockDAO := new(MockUsuarioDAO)
	svc := services.NuevoAuthService(mockDAO)

	dto := domain.DTORegistro{
		Nombre:   "María",
		Apellido: "García",
		Email:    "duplicado@example.com",
		Password: "password123",
	}

	// El email YA existe
	mockDAO.On("ExisteEmail", dto.Email).Return(true, nil)

	respuesta, err := svc.Registrar(dto)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
	assert.Contains(t, err.Error(), "ya está registrado")
	// Crear NO debe ser llamado si el email existe
	mockDAO.AssertNotCalled(t, "Crear", mock.Anything)
}

func TestRegistrar_ErrorAlVerificarEmail(t *testing.T) {
	mockDAO := new(MockUsuarioDAO)
	svc := services.NuevoAuthService(mockDAO)

	dto := domain.DTORegistro{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Error de base de datos al verificar
	mockDAO.On("ExisteEmail", dto.Email).Return(false, errors.New("error de conexión"))

	respuesta, err := svc.Registrar(dto)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
}

func TestRegistrar_ErrorAlCrearEnDB(t *testing.T) {
	mockDAO := new(MockUsuarioDAO)
	svc := services.NuevoAuthService(mockDAO)

	dto := domain.DTORegistro{
		Nombre:   "Pedro",
		Apellido: "Lopez",
		Email:    "pedro@example.com",
		Password: "password123",
	}

	mockDAO.On("ExisteEmail", dto.Email).Return(false, nil)
	// Simular error en la base de datos al crear
	mockDAO.On("Crear", mock.AnythingOfType("*domain.Usuario")).Return(errors.New("error de DB"))

	respuesta, err := svc.Registrar(dto)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
}

// ════════════════════════════════════════════════════════════════════
// Tests de Login
// ════════════════════════════════════════════════════════════════════

func TestLogin_ExitosoConCredencialesCorrectas(t *testing.T) {
	mockDAO := new(MockUsuarioDAO)
	svc := services.NuevoAuthService(mockDAO)

	// Hashear la contraseña como lo haría el sistema
	hash, _ := utils.HashearPassword("miPassword123")

	usuario := &domain.Usuario{
		ID:           1,
		Nombre:       "Ana",
		Apellido:     "Martinez",
		Email:        "ana@example.com",
		PasswordHash: hash,
		Rol:          domain.RolCliente,
		Activo:       true,
	}

	dto := domain.DTOLogin{
		Email:    "ana@example.com",
		Password: "miPassword123",
	}

	mockDAO.On("BuscarPorEmail", dto.Email).Return(usuario, nil)

	respuesta, err := svc.Login(dto)

	assert.NoError(t, err)
	assert.NotNil(t, respuesta)
	assert.NotEmpty(t, respuesta.Token)
	assert.Equal(t, usuario.Email, respuesta.Usuario.Email)
}

func TestLogin_ErrorConPasswordIncorrecto(t *testing.T) {
	mockDAO := new(MockUsuarioDAO)
	svc := services.NuevoAuthService(mockDAO)

	hash, _ := utils.HashearPassword("passwordCorrecto")
	usuario := &domain.Usuario{
		ID:           2,
		Email:        "test@example.com",
		PasswordHash: hash,
		Activo:       true,
	}

	dto := domain.DTOLogin{
		Email:    "test@example.com",
		Password: "passwordIncorrecto", // <-- contraseña errónea
	}

	mockDAO.On("BuscarPorEmail", dto.Email).Return(usuario, nil)

	respuesta, err := svc.Login(dto)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
	assert.Contains(t, err.Error(), "inválidas")
}

func TestLogin_ErrorEmailNoRegistrado(t *testing.T) {
	mockDAO := new(MockUsuarioDAO)
	svc := services.NuevoAuthService(mockDAO)

	dto := domain.DTOLogin{
		Email:    "noexiste@example.com",
		Password: "password123",
	}

	// Simular "record not found" de GORM
	mockDAO.On("BuscarPorEmail", dto.Email).Return(nil, errors.New("record not found"))

	respuesta, err := svc.Login(dto)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
}

func TestLogin_ErrorUsuarioInactivo(t *testing.T) {
	mockDAO := new(MockUsuarioDAO)
	svc := services.NuevoAuthService(mockDAO)

	hash, _ := utils.HashearPassword("password123")
	usuario := &domain.Usuario{
		ID:           3,
		Email:        "inactivo@example.com",
		PasswordHash: hash,
		Activo:       false, // <-- usuario desactivado
	}

	dto := domain.DTOLogin{
		Email:    "inactivo@example.com",
		Password: "password123",
	}

	mockDAO.On("BuscarPorEmail", dto.Email).Return(usuario, nil)

	respuesta, err := svc.Login(dto)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
	assert.Contains(t, err.Error(), "desactivada")
}
