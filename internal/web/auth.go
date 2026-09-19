package web

import (
	"crypto/subtle"
	"net/http"
)

// BasicAuth は全ルートに Basic 認証をかけるミドルウェア。
// user / pass が両方空のときは認証を無効化する (ローカル開発用。本番は必ず設定する)。
func BasicAuth(user, pass string, next http.Handler) http.Handler {
	enabled := user != "" || pass != ""
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if enabled {
			u, p, ok := r.BasicAuth()
			if !ok || !equal(u, user) || !equal(p, pass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="CardGame Admin"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// equal は定数時間比較 (タイミング攻撃対策)。
func equal(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
