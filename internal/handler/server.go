package handler

import (
	"log/slog"
	"net/http"

	"plown/internal/lobby"
	"plown/internal/repo"

	"github.com/go-chi/render"
)

type Handler struct {
	l             *slog.Logger
	repo          *repo.Postgres
	mgr           *lobby.Manager
	staticHandler http.Handler
}

func New(l *slog.Logger, repo *repo.Postgres, mgr *lobby.Manager, staticDir string) *Handler {
	return &Handler{
		l:             l,
		repo:          repo,
		mgr:           mgr,
		staticHandler: http.FileServer(http.Dir(staticDir)),
	}
}

func (h *Handler) Static(w http.ResponseWriter, r *http.Request) {
	h.staticHandler.ServeHTTP(w, r)
}

func (h *Handler) CreateLobby(w http.ResponseWriter, r *http.Request) {
	lobby, err := h.mgr.CreateLobby()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"lobby_id": lobby.ID})
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.repo.Tags(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rspTags := make([]Tag, 0, len(tags))
	for _, tag := range tags {
		rspTags = append(rspTags, Tag{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, rspTags)
}

func (h *Handler) GetLabels(w http.ResponseWriter, r *http.Request) {
	labels, err := h.repo.Labels(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, labels)
}
