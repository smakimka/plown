package lobby

import (
	"log/slog"
	"math/rand"
	"plown/internal/repo"
	"sync"
)

type Manager struct {
	mu   sync.RWMutex
	l    *slog.Logger
	repo *repo.Postgres

	lobbies map[string]*Lobby
}

func NewManager(l *slog.Logger, repo *repo.Postgres) *Manager {
	return &Manager{
		l:       l,
		repo:    repo,
		lobbies: make(map[string]*Lobby),
	}
}

func (m *Manager) CreateLobby() (*Lobby, error) {
	id := generateID(8)

	l, err := New(m.l.With("lobby_id", id), id, m.repo)
	if err != nil {
		m.l.Error("creating lobby", "error", err.Error())
		return nil, err
	}

	m.mu.Lock()
	m.lobbies[id] = l
	m.mu.Unlock()
	return l, nil
}

func (m *Manager) GetLobby(id string) *Lobby {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lobbies[id]
}

func generateID(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
