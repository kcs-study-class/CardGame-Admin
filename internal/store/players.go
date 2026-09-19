package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Player はプレイヤー一覧・詳細で表示する行。
type Player struct {
	ID        int64
	DeviceID  string
	Name      string
	Chips     int64
	Level     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ChipTransaction はチップ台帳の1行。
type ChipTransaction struct {
	ID         int64
	PlayerID   int64
	Delta      int64
	Reason     string
	RefMatchID *int64 // NULL 可
	CreatedAt  time.Time
}

// 管理操作のチップ増減理由。
const ReasonAdmin = "admin"

// ListPlayers はプレイヤーを検索して返す。query が空なら全件 (新しい順・上限 limit)。
// query は id / device_id / name の部分一致で絞り込む。
func (s *Store) ListPlayers(ctx context.Context, query string, limit int) ([]Player, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	var rows *sql.Rows
	var err error

	if query == "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, device_id, name, chips, level, created_at, updated_at
			   FROM players ORDER BY id DESC LIMIT ?`, limit)
	} else {
		like := "%" + query + "%"
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, device_id, name, chips, level, created_at, updated_at
			   FROM players
			  WHERE CAST(id AS CHAR) = ? OR device_id LIKE ? OR name LIKE ?
			  ORDER BY id DESC LIMIT ?`, query, like, like, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("store: list players: %w", err)
	}
	defer rows.Close()

	players := make([]Player, 0)
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.ID, &p.DeviceID, &p.Name, &p.Chips, &p.Level, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// GetPlayer は1人のプレイヤーを取得する。
func (s *Store) GetPlayer(ctx context.Context, id int64) (Player, error) {
	var p Player
	err := s.db.QueryRowContext(ctx,
		`SELECT id, device_id, name, chips, level, created_at, updated_at
		   FROM players WHERE id = ?`, id,
	).Scan(&p.ID, &p.DeviceID, &p.Name, &p.Chips, &p.Level, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return Player{}, err
	}
	return p, nil
}

// AdjustChips はチップ残高を delta 分増減し、同時に chip_transactions へ台帳を1行記録する
// (同一トランザクション)。残高キャッシュ (players.chips) と台帳を常に一致させる。
// これはゲームサーバーの store.AdjustChips と同じ設計 (裸 UPDATE を避ける)。
func (s *Store) AdjustChips(ctx context.Context, playerID int64, delta int64, reason string) (int64, error) {
	if reason == "" {
		reason = ReasonAdmin
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("store: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`UPDATE players SET chips = chips + ? WHERE id = ?`, delta, playerID); err != nil {
		return 0, fmt.Errorf("store: update chips: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO chip_transactions (player_id, delta, reason) VALUES (?, ?, ?)`,
		playerID, delta, reason); err != nil {
		return 0, fmt.Errorf("store: insert chip_transaction: %w", err)
	}

	var chips int64
	if err := tx.QueryRowContext(ctx,
		`SELECT chips FROM players WHERE id = ?`, playerID).Scan(&chips); err != nil {
		return 0, fmt.Errorf("store: reselect chips: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("store: commit: %w", err)
	}
	return chips, nil
}

// ListTransactions は指定プレイヤーのチップ台帳を新しい順に返す。
func (s *Store) ListTransactions(ctx context.Context, playerID int64, limit int) ([]ChipTransaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, player_id, delta, reason, ref_match_id, created_at
		   FROM chip_transactions WHERE player_id = ?
		  ORDER BY id DESC LIMIT ?`, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("store: list transactions: %w", err)
	}
	defer rows.Close()

	txns := make([]ChipTransaction, 0)
	for rows.Next() {
		var t ChipTransaction
		if err := rows.Scan(&t.ID, &t.PlayerID, &t.Delta, &t.Reason, &t.RefMatchID, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}
