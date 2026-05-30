package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerarCodigoQR genera un código único para identificar una entrada.
// El código combina timestamp + bytes aleatorios para garantizar unicidad.
func GenerarCodigoQR(eventoID, usuarioID uint) string {
	timestamp := time.Now().UnixNano()

	// 8 bytes aleatorios = 16 caracteres hex
	bytesAleatorios := make([]byte, 8)
	rand.Read(bytesAleatorios) //nolint:errcheck
	parteAleatoria := hex.EncodeToString(bytesAleatorios)

	return fmt.Sprintf("TKT-%d-%d-%d-%s", eventoID, usuarioID, timestamp, parteAleatoria)
}
