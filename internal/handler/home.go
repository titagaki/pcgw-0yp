package handler

import (
	"net/http"

	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/model"
)

func (h *Handler) TopPage(w http.ResponseWriter, r *http.Request) {
	if middleware.IsLoggedIn(r) {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}
	h.render(w, r, "top.html", nil)
}

func (h *Handler) DescPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "desc.html", nil)
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	channels, _ := model.ListChannelsByUser(h.DB, user.ID)
	recentInfos, _ := model.ListChannelInfosByUser(h.DB, user.ID, 10)

	data := map[string]interface{}{
		"Channels":    channels,
		"RecentInfos": recentInfos,
		"Flashes":     h.getFlashes(r, w),
	}
	h.render(w, r, "home.html", data)
}
