package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var data Data

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("can't read body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Printf("can't read body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = data.Validate()
	if err != nil {
		HandleValidationErr(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code, err := h.service.Shorten(data.Website)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(code))
}

func (h *Handler) Reroute(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	log.Printf("reroute called with code=%q", code)

	site, err := h.service.Reroute(code)
	if err != nil {
		log.Printf("err: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, site, http.StatusPermanentRedirect)
}
