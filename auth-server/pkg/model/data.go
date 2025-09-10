package model

import (
	"time"
)

// AuthUser representa a un usuario autenticado.
type AuthUser struct {
	// Identidad y credenciales
	Email        string    `json:"email"`          // Identificador único del usuario
	PasswordHash string    `json:"password_hash"`  // Hash de la contraseña (solo "local")
	Provider     string    `json:"provider"`       // "local", "google", "github", etc.
	// Autorización
	Role string `json:"role"` // "user", "admin", etc.
	// Metadatos
	CreatedAt time.Time `json:"created_at"` // Cuándo se registró
}


