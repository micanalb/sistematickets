package services

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"ticketsya/dao"
	"ticketsya/domain"
	"ticketsya/utils"
)

// AuthService define el contrato del servicio de autenticación
type AuthService interface {
	Registrar(dto domain.DTORegistro) (*domain.DTORespuestaLogin, error)
	Login(dto domain.DTOLogin) (*domain.DTORespuestaLogin, error)
}

// authServiceImpl es la implementación concreta
type authServiceImpl struct {
	usuarioDAO dao.UsuarioDAO
}

// NuevoAuthService crea una nueva instancia del servicio de autenticación
func NuevoAuthService(usuarioDAO dao.UsuarioDAO) AuthService {
	return &authServiceImpl{usuarioDAO: usuarioDAO}
}

// Registrar crea un nuevo usuario en el sistema.
// Valida que el email no esté registrado y hashea la contraseña antes de persistirla.
func (s *authServiceImpl) Registrar(dto domain.DTORegistro) (*domain.DTORespuestaLogin, error) {
	// Verificar que el email no exista previamente
	existe, err := s.usuarioDAO.ExisteEmail(dto.Email)
	if err != nil {
		return nil, fmt.Errorf("error al verificar email: %w", err)
	}
	if existe {
		return nil, errors.New("el email ya está registrado en el sistema")
	}

	// Hashear la contraseña ANTES de persistir - nunca guardar texto plano
	hash, err := utils.HashearPassword(dto.Password)
	if err != nil {
		return nil, fmt.Errorf("error al procesar contraseña: %w", err)
	}

	// Construir la entidad usuario
	usuario := &domain.Usuario{
		Nombre:       dto.Nombre,
		Apellido:     dto.Apellido,
		Email:        dto.Email,
		PasswordHash: hash,
		Telefono:     dto.Telefono,
		Rol:          domain.RolCliente, // Por defecto, todo registro es cliente
	}

	if err := s.usuarioDAO.Crear(usuario); err != nil {
		return nil, fmt.Errorf("error al crear usuario: %w", err)
	}

	// Generar token JWT para el nuevo usuario (login automático post-registro)
	token, err := utils.GenerarTokenJWT(usuario.ID, usuario.Email, usuario.Rol)
	if err != nil {
		return nil, fmt.Errorf("error al generar token: %w", err)
	}

	return &domain.DTORespuestaLogin{
		Token: token,
		Usuario: domain.UsuarioPublico{
			ID:       usuario.ID,
			Nombre:   usuario.Nombre,
			Apellido: usuario.Apellido,
			Email:    usuario.Email,
			Rol:      usuario.Rol,
			Telefono: usuario.Telefono,
		},
	}, nil
}

// Login autentica un usuario por email y contraseña.
// Verifica el hash y genera un token JWT si las credenciales son correctas.
func (s *authServiceImpl) Login(dto domain.DTOLogin) (*domain.DTORespuestaLogin, error) {
	// Buscar el usuario por email
	usuario, err := s.usuarioDAO.BuscarPorEmail(dto.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Mensaje genérico para no revelar si el email existe o no (seguridad)
			return nil, errors.New("credenciales inválidas")
		}
		return nil, fmt.Errorf("error al buscar usuario: %w", err)
	}

	// Verificar que el usuario esté activo
	if !usuario.Activo {
		return nil, errors.New("cuenta de usuario desactivada")
	}

	// Comparar contraseña ingresada con el hash almacenado
	if !utils.VerificarPassword(dto.Password, usuario.PasswordHash) {
		return nil, errors.New("credenciales inválidas")
	}

	// Generar token JWT firmado
	token, err := utils.GenerarTokenJWT(usuario.ID, usuario.Email, usuario.Rol)
	if err != nil {
		return nil, fmt.Errorf("error al generar token: %w", err)
	}

	return &domain.DTORespuestaLogin{
		Token: token,
		Usuario: domain.UsuarioPublico{
			ID:       usuario.ID,
			Nombre:   usuario.Nombre,
			Apellido: usuario.Apellido,
			Email:    usuario.Email,
			Rol:      usuario.Rol,
			Telefono: usuario.Telefono,
		},
	}, nil
}
