package handler

import (
	"net/http"
	"strconv"

	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/model"
	"github.com/titagaki/pcgw-0yp/internal/view/page"
)

func (h *Handler) SourceIndex(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	sources, _ := model.ListSourcesByUser(h.DB, user.ID)
	pd := h.pageData(r, w)
	h.renderTempl(w, r, page.Sources(pd, sources))
}

func (h *Handler) SourceAdd(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	name := r.URL.Query().Get("name")

	if name == "" {
		h.flash(w, r, "ソース名を指定してください")
		http.Redirect(w, r, "/sources", http.StatusFound)
		return
	}

	count, _ := model.CountSourcesByUser(h.DB, user.ID)
	if count >= 3 {
		h.flash(w, r, "ソースは最大3つまでです")
		http.Redirect(w, r, "/sources", http.StatusFound)
		return
	}

	model.CreateSource(h.DB, user.ID, name)
	h.flash(w, r, "ソースを追加しました")
	http.Redirect(w, r, "/sources", http.StatusFound)
}

func (h *Handler) SourceDelete(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Redirect(w, r, "/sources", http.StatusFound)
		return
	}

	source, err := model.GetSource(h.DB, id)
	if err != nil || source.UserID != user.ID {
		http.Redirect(w, r, "/sources", http.StatusFound)
		return
	}

	model.DeleteSource(h.DB, id)
	h.flash(w, r, "ソースを削除しました")
	http.Redirect(w, r, "/sources", http.StatusFound)
}

func (h *Handler) SourceRegen(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Redirect(w, r, "/sources", http.StatusFound)
		return
	}

	source, err := model.GetSource(h.DB, id)
	if err != nil || source.UserID != user.ID {
		http.Redirect(w, r, "/sources", http.StatusFound)
		return
	}

	model.RegenerateSourceKey(h.DB, id)
	h.flash(w, r, "キーを再生成しました")
	http.Redirect(w, r, "/sources", http.StatusFound)
}
