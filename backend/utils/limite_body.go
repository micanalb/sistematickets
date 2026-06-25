package utils

import "io"

// LimitarTamanioBody envuelve el body de una request con io.LimitReader,
// para que Gin/multipart no lea más bytes que el máximo permitido al
// procesar un archivo subido. Se usa antes de c.FormFile(...) en los
// endpoints de subida de imágenes, para rechazar archivos demasiado
// grandes sin tener que leerlos completos primero.
func LimitarTamanioBody(body io.ReadCloser, maxBytes int64) io.ReadCloser {
	return &cuerpoLimitado{io.LimitReader(body, maxBytes), body}
}

// cuerpoLimitado adapta un io.Reader limitado para que siga cumpliendo
// la interfaz io.ReadCloser (Gin espera poder cerrar el body original).
type cuerpoLimitado struct {
	io.Reader
	cerrable io.Closer
}

func (c *cuerpoLimitado) Close() error {
	return c.cerrable.Close()
}
