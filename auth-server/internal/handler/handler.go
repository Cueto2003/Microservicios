package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
	"fmt"
	"os"
	"io"
	"net/url"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"proyecto/auth-server/internal/controller"
	"proyecto/auth-server/internal/repository"
	"proyecto/auth-server/pkg/model"
)

//El puntero al controlador 
type Handler struct {
	ctrl *controller.Controller
}

//Es el constructor para creear el controlador 
func New(ctrl *controller.Controller) *Handler {
	return &Handler{ctrl}
}

func (h * Handler) RegisterUser(w http.ResponseWriter, req *http.Request) {
	//Falta checar si no ya hay un usuario que tenga el mismo correo 
	//obtener el password 
	Password := req.FormValue("password")
	//hacer el hash para despues guardarlo 
	Hash, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error al hashear la contraseña: %v", err)
		http.Error(w, "Error interno al registrar usuario", http.StatusInternalServerError)
		return
	}
	//Obtener el Id 
	User := model.AuthUser{
		Email:      	req.FormValue("email"),
		PasswordHash:	string(Hash), 
		Provider: 	 	req.FormValue("provider"),
		Role: 		 	req.FormValue("role"),
		CreatedAt:    	time.Now(),
	}
	
	if User.Email == "" {
		http.Error(w, "El campo 'email' es obligatorio.", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	//Obtiene los datos o el error 
	m, err := h.ctrl.Put(ctx, &User)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Repository error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// codifica y manda la espuesta 
	if err := json.NewEncoder(w).Encode(m); err != nil {
		log.Printf("Response error: v%\n", err)
	}

	//hacer la peticion al otro microservicio 
	
	// Preparar perfil
	profile := controller.MetadataUser{
		Email:       req.FormValue("email"),
		FullName:    req.FormValue("full_name"),
		AvatarURL:   req.FormValue("avatar_url"),
		PhoneNumber: req.FormValue("phone_number"),
		BirthDate:   req.FormValue("birth_date"),
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	// Enviar como formulario
	form := url.Values{}
	form.Add("email", profile.Email)
	form.Add("full_name", profile.FullName)
	form.Add("avatar_url", profile.AvatarURL)
	form.Add("phone_number", profile.PhoneNumber)
	form.Add("birth_date", profile.BirthDate)
	form.Add("last_updated", profile.LastUpdated)

	resp, err := http.PostForm("http://localhost:8081/MetadataUser", form)
	if err != nil {
		log.Printf("Error llamando a MetadataUser: %v", err)
		http.Error(w, "Error creando perfil", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Respuesta inesperada del microservicio Metadata: %s", body)
		http.Error(w, "Error al crear metadatos", http.StatusInternalServerError)
		return
	}

	// Respuesta final al cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "correct",
		"email":  profile.Email,
	})
}
// Va a resibir el correo con la contraseña 
func (h * Handler) Login(w http.ResponseWriter, req *http.Request) {

	Email := req.FormValue("email")
	Password := req.FormValue("password")


	if Email == "" || Password == "" {
		http.Error(w, "Email y password son obligatorios.", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
    // obtener el user deacuerdo al correo para despues hacer la comparación 
	User , err := h.ctrl.GetHashByEmail(ctx, Email)
	// Marcar error si no lo encuentra 
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "Usuario no encontrado", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	
	if !h.ctrl.CheckPasswordHash(Password, User.PasswordHash) {
		// Contraseña incorrecta
		log.Printf("Login fallido: contraseña inválida para el usuario %s", Email)
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	//Generar JTI (ID único del token)
    jti := uuid.NewString()
	
	// Para crearlo 
    secret := []byte(os.Getenv("JWT_SECRET"))
	// se genera el token (es un string largo)
	accessToken, err :=  h.ctrl.GenerateAccessToken(secret, User.Email, User.Email, User.Role, jti, 15*time.Minute)
	//maneja los errores 
    if err != nil {
        log.Printf("Error generando token: %v", err)
        http.Error(w, "Error interno", http.StatusInternalServerError)
        return
    }

    //Devolver token al cliente
    w.Header().Set("Content-Type", "application/json")
	//Para dar la respuesta en formato: 
	/*
		{
			"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9....",
			"token_type": "Bearer",
			"expires_in": "900"
		}
	*/
    json.NewEncoder(w).Encode(map[string]string{
        "access_token": accessToken,
        "token_type":   "Bearer",
        "expires_in":   fmt.Sprintf("%d", int((15 * time.Minute).Seconds())),
    })
}

/*func (h * Handler) VerifyToken(w http.ResponseWriter, req *http.Request) {

}*/
