package web

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/kcs-study-class/CardGame-Admin/internal/store"
)

// Routes は http.ServeMux にハンドラを登録して返す。
func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.players)
	mux.HandleFunc("GET /players/{id}", h.playerDetail)
	mux.HandleFunc("POST /players/{id}/adjust", h.adjustChips)
	mux.HandleFunc("GET /notices", h.notices)
	mux.HandleFunc("POST /notices", h.createNotice)
	mux.HandleFunc("POST /notices/{id}/delete", h.deleteNotice)
	mux.HandleFunc("GET /maintenance", h.maintenance)
	mux.HandleFunc("POST /maintenance", h.createMaintenance)
	mux.HandleFunc("POST /maintenance/{id}/delete", h.deleteMaintenance)
	return mux
}

// ---- プレイヤー ----

func (h *Handler) players(w http.ResponseWriter, r *http.Request) {
	// "/" 以外の未定義パスはここに来るので 404 にする
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	query := r.URL.Query().Get("q")
	players, err := h.store.ListPlayers(r.Context(), query, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, "players", "プレイヤー", r.URL.Query().Get("flash"), struct {
		Query   string
		Players []store.Player
	}{query, players})
}

func (h *Handler) playerDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(w, r)
	if !ok {
		return
	}
	player, err := h.store.GetPlayer(r.Context(), id)
	if err != nil {
		http.Error(w, "プレイヤーが見つかりません", http.StatusNotFound)
		return
	}
	txns, err := h.store.ListTransactions(r.Context(), id, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, "player_detail", "プレイヤー詳細", r.URL.Query().Get("flash"), struct {
		Player       store.Player
		Transactions []store.ChipTransaction
	}{player, txns})
}

func (h *Handler) adjustChips(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(w, r)
	if !ok {
		return
	}
	delta, err := strconv.ParseInt(r.FormValue("delta"), 10, 64)
	if err != nil {
		http.Error(w, "delta が不正です", http.StatusBadRequest)
		return
	}
	reason := r.FormValue("reason")
	chips, err := h.store.AdjustChips(r.Context(), id, delta, reason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirect(w, r, "/players/"+strconv.FormatInt(id, 10), "チップを調整しました (残高: "+strconv.FormatInt(chips, 10)+")")
}

// ---- お知らせ ----

func (h *Handler) notices(w http.ResponseWriter, r *http.Request) {
	notices, err := h.store.ListNotices(r.Context(), 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, "notices", "お知らせ", r.URL.Query().Get("flash"), struct {
		Notices []store.Notice
	}{notices})
}

func (h *Handler) createNotice(w http.ResponseWriter, r *http.Request) {
	startsAt, endsAt, err := parseRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateNotice(r.Context(), r.FormValue("title"), r.FormValue("body"), startsAt, endsAt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirect(w, r, "/notices", "お知らせを追加しました")
}

func (h *Handler) deleteNotice(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteNotice(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirect(w, r, "/notices", "削除しました")
}

// ---- メンテナンス ----

func (h *Handler) maintenance(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListMaintenance(r.Context(), 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, "maintenance", "メンテナンス", r.URL.Query().Get("flash"), struct {
		Items []store.Maintenance
	}{items})
}

func (h *Handler) createMaintenance(w http.ResponseWriter, r *http.Request) {
	startsAt, endsAt, err := parseRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateMaintenance(r.Context(), r.FormValue("message"), startsAt, endsAt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirect(w, r, "/maintenance", "メンテナンス予定を追加しました")
}

func (h *Handler) deleteMaintenance(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteMaintenance(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirect(w, r, "/maintenance", "削除しました")
}

// ---- ヘルパ ----

// pathInt は URL パスの {id} を int64 で取り出す。失敗時は 400 を返し false。
func pathInt(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id が不正です", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

// parseRange はフォームの startsAt / endsAt (datetime-local) をパースする。
// datetime-local はタイムゾーンを持たないため、ローカル時刻として解釈し UTC に変換する。
func parseRange(r *http.Request) (time.Time, time.Time, error) {
	const layout = "2006-01-02T15:04"
	startsAt, err := time.ParseInLocation(layout, r.FormValue("startsAt"), time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endsAt, err := time.ParseInLocation(layout, r.FormValue("endsAt"), time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return startsAt.UTC(), endsAt.UTC(), nil
}

// redirect は処理後に対象ページへ 303 リダイレクトし、flash メッセージをクエリで渡す。
func redirect(w http.ResponseWriter, r *http.Request, path, flash string) {
	target := path
	if flash != "" {
		target += "?flash=" + url.QueryEscape(flash)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}
