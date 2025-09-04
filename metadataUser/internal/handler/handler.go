package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"proyecto/metadataUser/internal/controller"
	"proyecto/metadataUser/internal/repository"
	model "proyecto/metadataUser/pkg"
)

//El puntero al controlador 
type Handler struct {
	ctrl *metadataUser.Controller
}

//Es el constructor para creear el controlador 
func New(ctrl *metadataUser.Controller) *Handler {
	return &Handler{ctrl}
}

func (h * Handler) GetMetadatUser(w http.ResponseWriter, req *http.Request) {
    //Obtener el Id 
    Email := req.FormValue("email")
    //Verifica que no venga vacio 
    if Email == "" {
        w.WriteHeader(http.StatusBadRequest)
		return
    }

    ctx := req.Context()
    //Obtiene los datos o el error 
    m, err := h.ctrl.Get(ctx, Email)
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
}

// Es el handler 
func (h * Handler) CreateMetadatUser(w http.ResponseWriter, req *http.Request) {
	//Obtener el Id 
	user := model.MetadataUser{
		Email:       req.FormValue("email"),
		FullName:    req.FormValue("full_name"),
		AvatarURL:   req.FormValue("avatar_url"),
		PhoneNumber: req.FormValue("phone_number"),
		BirthDate:   req.FormValue("birth_date"),
		LastUpdated: time.Now().Format(time.RFC3339), // Fecha actual en formato ISO
	}
	//Verifica que no venga vacio 
	if user.Email == "" {
		http.Error(w, "El campo 'email' es obligatorio.", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	//Obtiene los datos o el error 
	m, err := h.ctrl.Put(ctx, &user)
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
}