// CardGame-Admin エントリポイント。
//
// DB (cardgame) に直接アクセスする管理者ページ。生徒には配らない「提供物」。
// ゲームサーバー (CardGame-Server) とは別リポジトリ・別コンテナで、同じ MySQL を共有する。
//
// エンドポイント:
//   - GET /healthz            : 死活監視 (認証不要。コンテナ healthcheck 用)
//   - GET /                   : プレイヤー一覧・検索
//   - GET /players/{id}       : プレイヤー詳細 + チップ台帳
//   - POST /players/{id}/adjust : チップ調整 (台帳付き・同一Tx)
//   - GET/POST /notices, /maintenance : お知らせ・メンテナンス管理
//
// 認証: Basic 認証 (ADMIN_USER / ADMIN_PASS)。全管理ルートに適用 (/healthz は除く)。
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kcs-study-class/CardGame-Admin/internal/store"
	"github.com/kcs-study-class/CardGame-Admin/internal/web"
)

func main() {
	addr := envOr("ADMIN_ADDR", ":9090")
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN が未設定です (例: root:cardgame@tcp(127.0.0.1:3306)/cardgame?parseTime=true&loc=UTC)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	st, err := store.Open(ctx, dsn)
	if err != nil {
		log.Fatalf("DB 接続に失敗しました: %v", err)
	}
	defer st.Close()

	handler := web.NewHandler(st)

	user := os.Getenv("ADMIN_USER")
	pass := os.Getenv("ADMIN_PASS")
	if user == "" && pass == "" {
		log.Printf("[admin] 警告: ADMIN_USER/ADMIN_PASS 未設定のため認証なしで起動します (本番では必ず設定)")
	}

	// /healthz は認証の外、それ以外は Basic 認証でラップする。
	root := http.NewServeMux()
	root.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	root.Handle("/", web.BasicAuth(user, pass, handler.Routes()))

	log.Printf("CardGame-Admin listening on %s", addr)
	if err := http.ListenAndServe(addr, root); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
