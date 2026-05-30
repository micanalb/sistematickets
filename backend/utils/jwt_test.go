package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ticketsya/domain"
	"ticketsya/utils"
)

func TestGenerarTokenJWT_ExitosoConDatosValidos(t *testing.T) {
	token, err := utils.GenerarTokenJWT(1, "test@example.com", domain.RolCliente)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidarTokenJWT_TokenValido(t *testing.T) {
	// Generar un token primero
	token, _ := utils.GenerarTokenJWT(42, "usuario@example.com", domain.RolAdministrador)

	// Validarlo
	claims, err := utils.ValidarTokenJWT(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, uint(42), claims.UsuarioID)
	assert.Equal(t, "usuario@example.com", claims.Email)
	assert.Equal(t, domain.RolAdministrador, claims.Rol)
}

func TestValidarTokenJWT_TokenInvalido(t *testing.T) {
	claims, err := utils.ValidarTokenJWT("esto.no.es.un.token.valido")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidarTokenJWT_TokenVacio(t *testing.T) {
	claims, err := utils.ValidarTokenJWT("")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestHashearPassword_ProduceHashDistintoAlOriginal(t *testing.T) {
	password := "miPasswordSeguro123"
	hash, err := utils.HashearPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash) // No deben ser iguales
}

func TestVerificarPassword_PasswordCorrecta(t *testing.T) {
	password := "password123"
	hash, _ := utils.HashearPassword(password)

	resultado := utils.VerificarPassword(password, hash)

	assert.True(t, resultado)
}

func TestVerificarPassword_PasswordIncorrecta(t *testing.T) {
	hash, _ := utils.HashearPassword("passwordCorrecto")

	resultado := utils.VerificarPassword("passwordIncorrecto", hash)

	assert.False(t, resultado)
}

func TestHashearPassword_HashesDistintosParaMismaPassword(t *testing.T) {
	// bcrypt genera un salt diferente cada vez, por lo que dos hashes del mismo
	// password deben ser distintos (pero ambos verifican correctamente)
	password := "mismaPassword"
	hash1, _ := utils.HashearPassword(password)
	hash2, _ := utils.HashearPassword(password)

	assert.NotEqual(t, hash1, hash2)
	assert.True(t, utils.VerificarPassword(password, hash1))
	assert.True(t, utils.VerificarPassword(password, hash2))
}
