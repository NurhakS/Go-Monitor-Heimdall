package api

import (
	"encoding/json"
	"net/http"
	"uptime-monitor/services"

	"github.com/gorilla/mux"
)

type Handler struct {
	services *services.Services
}

func NewHandler(services *services.Services) *Handler {
	return &Handler{services: services}
}

func (h *Handler) RegisterCredentialsRoutes(r *mux.Router) {
	r.HandleFunc("/api/credentials", h.GetCredentials).Methods("GET")
	r.HandleFunc("/api/credentials", h.CreateCredential).Methods("POST")
	r.HandleFunc("/api/credentials/{id}", h.GetCredential).Methods("GET")
	r.HandleFunc("/api/credentials/{id}", h.UpdateCredential).Methods("PUT")
	r.HandleFunc("/api/credentials/{id}", h.DeleteCredential).Methods("DELETE")
}

func (h *Handler) GetCredentials(w http.ResponseWriter, r *http.Request) {
	profileID := r.Header.Get("X-Profile-ID")
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	credentials, err := h.services.Credentials.GetCredentials(profileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(credentials)
}

func (h *Handler) CreateCredential(w http.ResponseWriter, r *http.Request) {
	profileID := r.Header.Get("X-Profile-ID")
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	var cred services.Credential
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cred.ProfileID = profileID

	if err := h.services.Credentials.CreateCredential(&cred); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cred)
}

func (h *Handler) GetCredential(w http.ResponseWriter, r *http.Request) {
	profileID := r.Header.Get("X-Profile-ID")
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]
	cred, err := h.services.Credentials.GetCredential(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if cred.ProfileID != profileID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(cred)
}

func (h *Handler) UpdateCredential(w http.ResponseWriter, r *http.Request) {
	profileID := r.Header.Get("X-Profile-ID")
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]
	var cred services.Credential
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cred.ID = id
	cred.ProfileID = profileID

	if err := h.services.Credentials.UpdateCredential(&cred); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cred)
}

func (h *Handler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	profileID := r.Header.Get("X-Profile-ID")
	if profileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]
	if err := h.services.Credentials.DeleteCredential(id, profileID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
