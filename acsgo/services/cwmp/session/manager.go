// Package session manages TR-069 device session state.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// State tracks the CWMP session state machine for a device.
type State string

const (
	StateIdle      State = "idle"
	StateInform    State = "inform"
	StateWaiting   State = "waiting"  // waiting for CPE response to an ACS RPC
	StateComplete  State = "complete"
)

// Session holds the in-flight state for one CPE session.
type Session struct {
	DeviceID        string    `json:"device_id"`
	State           State     `json:"state"`
	CurrentCmdID    string    `json:"current_cmd_id"`
	CWMPID          string    `json:"cwmp_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PendingRPCCount int       `json:"pending_rpc_count"`
}

// Manager handles in-memory session lifecycle with Redis backing.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	cache    SessionCache
}

// SessionCache defines the minimal interface needed from the Redis cache.
type SessionCache interface {
	SetSession(ctx context.Context, deviceID string, data interface{}) error
	GetSession(ctx context.Context, deviceID string, dst interface{}) error
	DeleteSession(ctx context.Context, deviceID string) error
}

// NewManager creates a new session Manager.
func NewManager(cache SessionCache) *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		cache:    cache,
	}
}

// Begin starts a new session for a device.
func (m *Manager) Begin(ctx context.Context, deviceID, cwmpID string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := &Session{
		DeviceID:  deviceID,
		State:     StateInform,
		CWMPID:    cwmpID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.sessions[deviceID] = s
	_ = m.cache.SetSession(ctx, deviceID, s) // best-effort
	return s
}

// Get retrieves the current session for a device.
func (m *Manager) Get(ctx context.Context, deviceID string) (*Session, bool) {
	m.mu.RLock()
	s, ok := m.sessions[deviceID]
	m.mu.RUnlock()
	if ok {
		return s, true
	}

	// Try Redis fallback
	var rs Session
	if err := m.cache.GetSession(ctx, deviceID, &rs); err == nil {
		m.mu.Lock()
		m.sessions[deviceID] = &rs
		m.mu.Unlock()
		return &rs, true
	}
	return nil, false
}

// SetWaiting transitions a session to the "waiting for CPE response" state.
func (m *Manager) SetWaiting(ctx context.Context, deviceID, cmdID string) error {
	return m.transition(ctx, deviceID, StateWaiting, func(s *Session) {
		s.CurrentCmdID = cmdID
	})
}

// SetIdle transitions a session back to idle (no pending commands).
func (m *Manager) SetIdle(ctx context.Context, deviceID string) error {
	return m.transition(ctx, deviceID, StateIdle, nil)
}

// End removes a session after the CPE connection closes.
func (m *Manager) End(ctx context.Context, deviceID string) {
	m.mu.Lock()
	delete(m.sessions, deviceID)
	m.mu.Unlock()
	_ = m.cache.DeleteSession(ctx, deviceID)
}

// CurrentCmdID returns the command ID being waited on, or empty if none.
func (m *Manager) CurrentCmdID(ctx context.Context, deviceID string) string {
	s, ok := m.Get(ctx, deviceID)
	if !ok {
		return ""
	}
	return s.CurrentCmdID
}

func (m *Manager) transition(ctx context.Context, deviceID string, state State, fn func(*Session)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[deviceID]
	if !ok {
		return fmt.Errorf("session not found for device %s", deviceID)
	}
	s.State = state
	s.UpdatedAt = time.Now()
	if fn != nil {
		fn(s)
	}

	// Best-effort cache update
	b, _ := json.Marshal(s)
	_ = b
	_ = m.cache.SetSession(ctx, deviceID, s)
	return nil
}
