// Package web は管理ページの HTTP ハンドラと画面描画を担う。
package web

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/kcs-study-class/CardGame-Admin/internal/store"
)

//go:embed templates/*.html
var templateFS embed.FS

// pageData はレイアウトに渡す共通データ。Data に各ページ固有の値を入れる。
type pageData struct {
	Title string
	Flash string
	Data  any
}

// render は layout.html と指定ページを組み合わせて描画する。
// page は "players" のようにファイル名 (拡張子なし) を渡す。
func render(w http.ResponseWriter, page, title, flash string, data any) {
	tmpl, err := template.ParseFS(templateFS, "templates/layout.html", "templates/"+page+".html")
	if err != nil {
		log.Printf("[web] template parse: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", pageData{Title: title, Flash: flash, Data: data}); err != nil {
		log.Printf("[web] template exec: %v", err)
	}
}

// Handler は管理ページのルーティングを組み立てて返す。
type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}
