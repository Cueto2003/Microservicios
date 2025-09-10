package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"proyecto/auth-server/internal/controller"
	"proyecto/auth-server/internal/repository"
	"proyecto/auth-server/pkg/model"
	registry "proyecto/pkg/registry" // <-- usa este nombre 'registry'
)

type Handler struct {
	ctrl     *controller.Controller
	registry registry.Registry
}

func New(ctrl *controller.Controller, reg registry.Registry) *Handler {
	return &Handler{ctrl: ctrl, registry: reg}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	password := req.FormValue("password")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error al hashear la contraseña: %v", err)
		http.Error(w, "Error interno al registrar usuario", http.StatusInternalServerError)
		return
	}

	user := model.AuthUser{
		Email:        req.FormValue("email"),
		PasswordHash: string(hash),
		Provider:     req.FormValue("provider"),
		Role:         req.FormValue("role"),
		CreatedAt:    time.Now(),
	}
	if user.Email == "" {
		http.Error(w, "El campo 'email' es obligatorio.", http.StatusBadRequest)
		return
	}

	createdUser, err := h.ctrl.Put(ctx, &user)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "No encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Repository error: %v\n", err)
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	profile := controller.MetadataUser{
		Email:       req.FormValue("email"),
		FullName:    req.FormValue("full_name"),
		AvatarURL:   req.FormValue("avatar_url"),
		PhoneNumber: req.FormValue("phone_number"),
		BirthDate:   req.FormValue("birth_date"),
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	form := url.Values{}
	form.Add("email", profile.Email)
	form.Add("full_name", profile.FullName)
	form.Add("avatar_url", profile.AvatarURL)
	form.Add("phone_number", profile.PhoneNumber)
	form.Add("birth_date", profile.BirthDate)
	form.Add("last_updated", profile.LastUpdated)

	addrs, err := h.registry.ServiceAddress(ctx, "metadata-user")
	if err != nil {
		log.Printf("Service discovery error: %v", err)
		http.Error(w, "No se pudo resolver metadata-user", http.StatusBadGateway)
		return
	}
	target := "http://" + addrs[rand.Intn(len(addrs))] + "/MetadataUser"

	reqOut, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(form.Encode()))
	if err != nil {
		log.Printf("Error creando request a metadata-user: %v", err)
		http.Error(w, "Error creando petición a metadata", http.StatusInternalServerError)
		return
	}
	reqOut.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(reqOut)
	if err != nil {
		log.Printf("Error llamando a metadata-user (%s): %v", target, err)
		http.Error(w, "Error comunicando con metadata", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Respuesta inesperada de metadata-user (%d): %s", resp.StatusCode, string(body))
		http.Error(w, "Error al crear metadatos", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "correct",
		"auth_user": createdUser,
		"email":     profile.Email,
	})
}

func (h *Handler) Login(w http.ResponseWriter, req *http.Request) {
	email := req.FormValue("email")
	password := req.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Email y password son obligatorios.", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	user, err := h.ctrl.GetHashByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "Usuario no encontrado", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	if !h.ctrl.CheckPasswordHash(password, user.PasswordHash) {
		log.Printf("Login fallido: contraseña inválida para el usuario %s", email)
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	jti := uuid.NewString()
	secret := []byte(os.Getenv("JWT_SECRET"))

	accessToken, err := h.ctrl.GenerateAccessToken(secret, user.Email, user.Email, user.Role, jti, 15*time.Minute)
	if err != nil {
		log.Printf("Error generando token: %v", err)
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   fmt.Sprintf("%d", int((15 * time.Minute).Seconds())),
	})
}
