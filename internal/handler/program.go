package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/titagaki/pcgw-0yp/internal/model"
)

func (h *Handler) ProgramIndex(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := map[string]interface{}{
		"Year":  now.Year(),
		"Month": int(now.Month()),
	}
	h.render(w, r, "programs.html", data)
}

func (h *Handler) ProgramRecent(w http.ResponseWriter, r *http.Request) {
	infos, _ := model.ListRecentChannelInfos(h.DB, 20)
	data := map[string]interface{}{
		"Programs": infos,
	}
	h.render(w, r, "programs_recent.html", data)
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

	// Group by day
	byDay := make(map[int][]*model.ChannelInfo)
	for _, ci := range infos {
		day := ci.CreatedAt.Day()
		byDay[day] = append(byDay[day], ci)
	}

	data := map[string]interface{}{
		"Year":     year,
		"Month":    month,
		"Programs": infos,
		"ByDay":    byDay,
	}
	h.render(w, r, "program_month.html", data)
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

	data := map[string]interface{}{
		"Program":     ci,
		"ProgramUser": user,
	}
	h.render(w, r, "program.html", data)
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

	if !ci.TerminatedAt.Valid {
		h.flash(w, r, "配信中の番組は削除できません")
		http.Redirect(w, r, "/programs", http.StatusFound)
		return
	}

	model.DeleteChannelInfo(h.DB, id)
	h.flash(w, r, "番組を削除しました")
	http.Redirect(w, r, "/programs", http.StatusFound)
}
