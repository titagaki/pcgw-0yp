package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/model"
	programview "github.com/titagaki/pcgw-0yp/internal/view/program"
)

func (h *Handler) ProgramIndex(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	pd := h.pageData(r, w)
	h.renderTempl(w, r, programview.Index(pd, now.Year(), int(now.Month())))
}

func (h *Handler) ProgramRecent(w http.ResponseWriter, r *http.Request) {
	infos, _ := model.ListRecentChannelInfos(h.DB, 20)
	pd := h.pageData(r, w)
	h.renderTempl(w, r, programview.Recent(pd, infos))
}

func (h *Handler) ProgramByMonth(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	month, err := strconv.Atoi(chi.URLParam(r, "month"))
	if err != nil || month < 1 || month > 12 {
		http.NotFound(w, r)
		return
	}

	infos, _ := model.ListChannelInfosByMonth(h.DB, year, month)

	pd := h.pageData(r, w)
	h.renderTempl(w, r, programview.Month(pd, year, month, infos))
}

func (h *Handler) ProgramShow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ci, err := model.GetChannelInfo(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user, _ := model.GetUser(h.DB, ci.UserID)

	pd := h.pageData(r, w)
	h.renderTempl(w, r, programview.Show(pd, ci, user))
}

func (h *Handler) ProgramDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ci, err := model.GetChannelInfo(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.CurrentUser(r)
	if user == nil || (user.ID != ci.UserID && !user.Admin) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if !ci.TerminatedAt.Valid {
		h.flash(w, r, "配信中の番組は削除できません")
		http.Redirect(w, r, "/programs", http.StatusFound)
		return
	}

	model.DeleteChannelInfo(h.DB, id)
	h.flash(w, r, "番組を削除しました")
	http.Redirect(w, r, "/programs", http.StatusFound)
}
