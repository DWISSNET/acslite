package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DWISSNET/acsgo/models"
	"github.com/DWISSNET/acsgo/pkg"
)

// RPCService manages the RPC command queue
type RPCService struct {
	db *sql.DB
}

func NewRPCService(db *sql.DB) *RPCService {
	return &RPCService{db: db}
}

// Enqueue adds a new RPC command for a device.
func (s *RPCService) Enqueue(serial, cmdType string, params models.JSONMap) (*models.RPCQueue, error) {
	id := pkg.NewUUID()
	p, err := json.Marshal(params)
	if err != nil {
		p = []byte("{}")
	}
	now := time.Now()
	_, err = s.db.Exec(`INSERT INTO rpc_queue (id, device_serial, command_type, parameters, status, created_at)
		VALUES (?,?,?,?,?,?)`, id, serial, cmdType, string(p), "pending", now)
	if err != nil {
		return nil, err
	}
	return &models.RPCQueue{
		ID:           id,
		DeviceSerial: serial,
		CommandType:  cmdType,
		Parameters:   params,
		Status:       "pending",
		CreatedAt:    now,
	}, nil
}

// NextPending returns the oldest pending command for a device (FIFO).
func (s *RPCService) NextPending(serial string) (*models.RPCQueue, error) {
	row := s.db.QueryRow(`SELECT id, device_serial, command_type, parameters, status, created_at
		FROM rpc_queue WHERE device_serial = ? AND status = 'pending'
		ORDER BY created_at ASC LIMIT 1`, serial)

	var q models.RPCQueue
	var paramsJSON string
	err := row.Scan(&q.ID, &q.DeviceSerial, &q.CommandType, &paramsJSON, &q.Status, &q.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(paramsJSON), &q.Parameters); err != nil {
		q.Parameters = models.JSONMap{}
	}
	return &q, nil
}

// MarkSent transitions a command to "sent".
func (s *RPCService) MarkSent(id string) error {
	now := time.Now()
	_, err := s.db.Exec(`UPDATE rpc_queue SET status = 'sent', delivered_at = ? WHERE id = ?`, now, id)
	return err
}

// MarkCompleted transitions a command to "completed".
func (s *RPCService) MarkCompleted(id string) error {
	_, err := s.db.Exec(`UPDATE rpc_queue SET status = 'completed' WHERE id = ?`, id)
	return err
}

// MarkFailed transitions a command to "failed" with an error message.
func (s *RPCService) MarkFailed(id, errMsg string) error {
	_, err := s.db.Exec(`UPDATE rpc_queue SET status = 'failed', error_message = ? WHERE id = ?`, errMsg, id)
	return err
}

// ListForDevice returns all RPC entries for a given serial (latest first).
func (s *RPCService) ListForDevice(serial string) ([]models.RPCQueue, error) {
	rows, err := s.db.Query(`SELECT id, device_serial, command_type, parameters, status, created_at, delivered_at, error_message
		FROM rpc_queue WHERE device_serial = ? ORDER BY created_at DESC LIMIT 100`, serial)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.RPCQueue
	for rows.Next() {
		var q models.RPCQueue
		var paramsJSON string
		var deliveredAt sql.NullTime
		err := rows.Scan(&q.ID, &q.DeviceSerial, &q.CommandType, &paramsJSON, &q.Status,
			&q.CreatedAt, &deliveredAt, &q.ErrorMessage)
		if err != nil {
			return nil, fmt.Errorf("scan rpc_queue: %w", err)
		}
		if deliveredAt.Valid {
			q.DeliveredAt = &deliveredAt.Time
		}
		if err := json.Unmarshal([]byte(paramsJSON), &q.Parameters); err != nil {
			q.Parameters = models.JSONMap{}
		}
		result = append(result, q)
	}
	return result, nil
}
