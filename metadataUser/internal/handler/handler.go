package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"proyecto/metadataUser/internal/controller"
	"proyecto/metadataUser/internal/repository"
)

//El puntero al controlador 
type Handler struct {
	ctrl *metadataUser.Controller
}

//Es el constructor para creear el controlador 
func New(ctrl *metadataUser.Controller) *Handler {
	return &Handler{ctrl}
}

// Es el handler 
func (h * Handler) GetMetadatUser(w http.ResponseWriter, req *http.Request) {
	//Obtener el Id 
	id := req.FormValue("id")
	//Verifica que no venga vacio 
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	ctx := req.Context()
	//Obtiene los datos o el error 
	m, err := h.ctrl.Get(ctx, id)
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