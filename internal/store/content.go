package store

import (
	"context"
	"fmt"
	"time"
)

// Notice はお知らせ1件。
type Notice struct {
	ID       int64
	Title    string
	Body     string
	StartsAt time.Time
	EndsAt   time.Time
}

// Maintenance はメンテナンス予定1件。
type Maintenance struct {
	ID       int64
	Message  string
	StartsAt time.Time
	EndsAt   time.Time
}

// ---- お知らせ ----

// ListNotices は全お知らせを新しい順に返す (管理用なので期間で絞らない)。
func (s *Store) ListNotices(ctx context.Context, limit int) ([]Notice, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, body, starts_at, ends_at FROM notices ORDER BY starts_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("store: list notices: %w", err)
	}
	defer rows.Close()

	notices := make([]Notice, 0)
	for rows.Next() {
		var n Notice
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.StartsAt, &n.EndsAt); err != nil {
			return nil, err
		}
		notices = append(notices, n)
	}
	return notices, rows.Err()
}

// CreateNotice はお知らせを1件追加する。
func (s *Store) CreateNotice(ctx context.Context, title, body string, startsAt, endsAt time.Time) error {
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO notices (title, body, starts_at, ends_at) VALUES (?, ?, ?, ?)`,
		title, body, startsAt.UTC(), endsAt.UTC()); err != nil {
		return fmt.Errorf("store: create notice: %w", err)
	}
	return nil
}

// DeleteNotice はお知らせを削除する。
func (s *Store) DeleteNotice(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM notices WHERE id = ?`, id); err != nil {
		return fmt.Errorf("store: delete notice: %w", err)
	}
	return nil
}

// ---- メンテナンス ----

// ListMaintenance は全メンテナンス予定を新しい順に返す。
func (s *Store) ListMaintenance(ctx context.Context, limit int) ([]Maintenance, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, message, starts_at, ends_at FROM maintenance_windows ORDER BY starts_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("store: list maintenance: %w", err)
	}
	defer rows.Close()

	items := make([]Maintenance, 0)
	for rows.Next() {
		var m Maintenance
		if err := rows.Scan(&m.ID, &m.Message, &m.StartsAt, &m.EndsAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

// CreateMaintenance はメンテナンス予定を1件追加する。
func (s *Store) CreateMaintenance(ctx context.Context, message string, startsAt, endsAt time.Time) error {
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO maintenance_windows (message, starts_at, ends_at) VALUES (?, ?, ?)`,
		message, startsAt.UTC(), endsAt.UTC()); err != nil {
		return fmt.Errorf("store: create maintenance: %w", err)
	}
	return nil
}

// DeleteMaintenance はメンテナンス予定を削除する。
func (s *Store) DeleteMaintenance(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM maintenance_windows WHERE id = ?`, id); err != nil {
		return fmt.Errorf("store: delete maintenance: %w", err)
	}
	return nil
}
