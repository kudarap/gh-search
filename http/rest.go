package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kudarap/ghsearch"
)

const contentType = "application/json; charset=utf-8"

type RestHandler struct {
	userSvc ghsearch.UserService
}

func (h *RestHandler) GETUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		usernames := r.URL.Query().Get("usernames")
		if strings.TrimSpace(usernames) == "" {
			encodeJSONResp(w, make([]struct{}, 0), http.StatusOK)
			return
		}

		splits := strings.Split(usernames, ",")
		users, err := h.userSvc.Users(ctx, splits)
		if err != nil {
			encodeJSONError(w, err, http.StatusBadRequest)
			return
		}

		encodeJSONResp(w, users, http.StatusOK)
	}
}

func NewRestHandler(us ghsearch.UserService) *RestHandler {
	return &RestHandler{us}
}

func encodeJSONResp(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func encodeJSONError(w http.ResponseWriter, err error, code int) {
	m := struct {
		Error string `json:"error"`
	}{err.Error()}
	encodeJSONResp(w, m, code)
}
