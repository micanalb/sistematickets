package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashearPassword aplica bcrypt con costo 12 a la contraseña en texto plano.
// bcrypt es más seguro que MD5 o SHA256 para contraseñas porque es
// computacionalmente costoso e incluye un salt automático.
// Decisión de diseño: usamos bcrypt en lugar de MD5/SHA256 porque fue
// diseñado específicamente para hashear contraseñas de forma segura.
func HashearPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerificarPassword compara una contraseña en texto plano contra su hash.
// Retorna true si coinciden, false en caso contrario.
func VerificarPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
