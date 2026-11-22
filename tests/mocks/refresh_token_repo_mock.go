package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/bonarizki/gin-boilerplate/internal/domain/models"
	"github.com/bonarizki/gin-boilerplate/internal/domain/repositories"
)

// MockRefreshTokenRepository is an in-memory RefreshTokenRepository for unit tests.
type MockRefreshTokenRepository struct {
	mu     sync.RWMutex
	nextID uint
	// key: token hash
	byHash map[string]*models.RefreshToken
}

// NewMockRefreshTokenRepository returns a new MockRefreshTokenRepository.
func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{byHash: make(map[string]*models.RefreshToken)}
}

// Ensure MockRefreshTokenRepository implements repositories.RefreshTokenRepository.
var _ repositories.RefreshTokenRepository = (*MockRefreshTokenRepository)(nil)

// Seed inserts a token record directly, for setting up "already existing" test state.
func (m *MockRefreshTokenRepository) Seed(token *models.RefreshToken) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byHash == nil {
		m.byHash = make(map[string]*models.RefreshToken)
	}
	if token.ID == 0 {
		m.nextID++
		token.ID = m.nextID
	}
	m.byHash[token.TokenHash] = token
}

func (m *MockRefreshTokenRepository) Create(_ context.Context, token *models.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byHash == nil {
		m.byHash = make(map[string]*models.RefreshToken)
	}
	m.nextID++
	token.ID = m.nextID
	cp := *token
	m.byHash[token.TokenHash] = &cp
	return nil
}

func (m *MockRefreshTokenRepository) GetByTokenHash(_ context.Context, hash string) (*models.RefreshToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.byHash[hash]; ok {
		cp := *t
		return &cp, nil
	}
	return nil, nil
}

func (m *MockRefreshTokenRepository) Rotate(_ context.Context, oldID uint, newToken *models.RefreshToken) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var old *models.RefreshToken
	for _, t := range m.byHash {
		if t.ID == oldID {
			old = t
			break
		}
	}
	if old == nil || old.RevokedAt != nil {
		return false, nil
	}
	now := time.Now()
	old.RevokedAt = &now
	old.ReplacedByHash = newToken.TokenHash
	if newToken.ID == 0 {
		m.nextID++
		newToken.ID = m.nextID
	}
	cp := *newToken
	m.byHash[newToken.TokenHash] = &cp
	return true, nil
}

func (m *MockRefreshTokenRepository) MarkRotated(_ context.Context, id uint, newTokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.byHash {
		if t.ID == id && t.RevokedAt == nil {
			now := time.Now()
			t.RevokedAt = &now
			t.ReplacedByHash = newTokenHash
			return nil
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) RevokeFamily(_ context.Context, familyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for _, t := range m.byHash {
		if t.FamilyID == familyID && t.RevokedAt == nil {
			t.RevokedAt = &now
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteExpiredBefore(_ context.Context, cutoff time.Time) (int64, error) {
	_ = cutoff
	return 0, nil
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(_ context.Context, userID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for _, t := range m.byHash {
		if t.UserID == userID && t.RevokedAt == nil {
			t.RevokedAt = &now
		}
	}
	return nil
}
