// Package repository provides PostgreSQL data access for commands.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/DWISSNET/acsgo/pkg/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CommandRepository manages RPC command persistence.
type CommandRepository struct {
	db *pgxpool.Pool
}

// NewCommandRepository creates a new CommandRepository.
func NewCommandRepository(db *pgxpool.Pool) *CommandRepository {
	return &CommandRepository{db: db}
}

// Create inserts a new command into the queue.
func (r *CommandRepository) Create(ctx context.Context, cmd *models.Command) error {
	q := `
INSERT INTO commands (
    id, device_id, command_type, parameters, status,
    max_retries, created_by, scheduled_at, expires_at, created_at, updated_at
) VALUES (
    uuid_generate_v4(),
    (SELECT id FROM devices WHERE device_id = $1),
    $2, $3::jsonb, $4, $5, $6::uuid, $7, $8, NOW(), NOW()
)
RETURNING id::text`

	params, err := marshalJSON(cmd.Parameters)
	if err != nil {
		return fmt.Errorf("marshal command params: %w", err)
	}

	return r.db.QueryRow(ctx, q,
		cmd.DeviceID, cmd.CommandType, params, cmd.Status,
		cmd.MaxRetries, nullableStr(cmd.CreatedBy),
		cmd.ScheduledAt, cmd.ExpiresAt,
	).Scan(&cmd.ID)
}

// GetPendingForDevice returns pending/queued commands for a specific device.
func (r *CommandRepository) GetPendingForDevice(ctx context.Context, deviceID string) ([]*models.Command, error) {
	q := `
SELECT c.id::text, d.device_id, c.command_type, c.parameters::text,
       c.status, c.retry_count, c.max_retries,
       COALESCE(c.created_by::text, ''), c.response, c.error_msg,
       c.scheduled_at, c.delivered_at, c.completed_at, c.expires_at,
       c.created_at, c.updated_at
FROM commands c
JOIN devices d ON d.id = c.device_id
WHERE d.device_id = $1
  AND c.status IN ('pending', 'queued')
  AND (c.expires_at IS NULL OR c.expires_at > NOW())
  AND (c.scheduled_at IS NULL OR c.scheduled_at <= NOW())
ORDER BY c.created_at ASC`

	rows, err := r.db.Query(ctx, q, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCommands(rows)
}

// UpdateStatus changes the status of a command.
func (r *CommandRepository) UpdateStatus(ctx context.Context, id string, status models.CommandStatus, errMsg string) error {
	now := time.Now()
	var q string
	switch status {
	case models.CommandDelivered:
		q = `UPDATE commands SET status=$1, error_msg=$2, delivered_at=$3, updated_at=NOW() WHERE id=$4::uuid`
	case models.CommandCompleted:
		q = `UPDATE commands SET status=$1, error_msg=$2, completed_at=$3, updated_at=NOW() WHERE id=$4::uuid`
	default:
		q = `UPDATE commands SET status=$1, error_msg=$2, updated_at=NOW() WHERE id=$3::uuid AND $4::timestamptz IS NOT NULL`
	}
	_, err := r.db.Exec(ctx, q, status, errMsg, now, id)
	return err
}

// IncrementRetry bumps the retry count and sets status to pending.
func (r *CommandRepository) IncrementRetry(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
UPDATE commands
SET retry_count = retry_count + 1,
    status = CASE WHEN retry_count + 1 >= max_retries THEN 'failed' ELSE 'pending' END,
    updated_at = NOW()
WHERE id = $1::uuid`, id)
	return err
}

// ExpireOld marks commands past their deadline as expired.
func (r *CommandRepository) ExpireOld(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `
UPDATE commands SET status='expired', updated_at=NOW()
WHERE status IN ('pending','queued') AND expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// scanCommands reads command rows.
func scanCommands(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]*models.Command, error) {
	var cmds []*models.Command
	for rows.Next() {
		c := &models.Command{}
		var paramsJSON string
		err := rows.Scan(
			&c.ID, &c.DeviceID, &c.CommandType, &paramsJSON,
			&c.Status, &c.RetryCount, &c.MaxRetries,
			&c.CreatedBy, &c.Response, &c.ErrorMsg,
			&c.ScheduledAt, &c.DeliveredAt, &c.CompletedAt, &c.ExpiresAt,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan command: %w", err)
		}
		if err := unmarshalJSON(paramsJSON, &c.Parameters); err != nil {
			c.Parameters = map[string]string{}
		}
		cmds = append(cmds, c)
	}
	return cmds, nil
}
