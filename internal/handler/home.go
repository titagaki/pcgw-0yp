package handler

import (
	"net/http"

	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/model"
	"github.com/titagaki/pcgw-0yp/internal/view/page"
)

func (h *Handler) TopPage(w http.ResponseWriter, r *http.Request) {
	if middleware.IsLoggedIn(r) {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}
	pd := h.pageData(r, w)
	h.renderTempl(w, r, page.Top(pd))
}

func (h *Handler) DescPage(w http.ResponseWriter, r *http.Request) {
	pd := h.pageData(r, w)
	h.renderTempl(w, r, page.Desc(pd))
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	channels, _ := model.ListChannelsByUser(h.DB, user.ID)
	recentInfos, _ := model.ListChannelInfosByUser(h.DB, user.ID, 10)

	pd := h.pageData(r, w)
	h.renderTempl(w, r, page.Home(pd, channels, recentInfos))
}
