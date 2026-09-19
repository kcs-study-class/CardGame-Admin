// Package store は管理ページから cardgame データベースへ直接アクセスする層。
// スキーマの正は CardGame-Server/Docs/DB.md。管理ページはマイグレーションしない
// (テーブル作成はゲームサーバー側の責務)。時刻はすべて UTC で扱う。
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Store は cardgame DB への接続を保持する。
type Store struct {
	db *sql.DB
}

// Open は DSN で接続し、疎通確認 (Ping) まで行う。
func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: sql.Open: %w", err)
	}
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }
