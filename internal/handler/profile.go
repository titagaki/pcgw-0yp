package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/repository"
	profileview "github.com/titagaki/pcgw-0yp/internal/view/profile"
)

func (h *Handler) ProfileList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query != "" {
		users, _ := repository.SearchUsersByName(h.DB, query)
		pd := h.pageData(r, w)
		h.renderTempl(w, r, profileview.List(pd, users, query))
	} else {
		users, _ := repository.ListActiveUsers(h.DB, 30)
		pd := h.pageData(r, w)
		h.renderTempl(w, r, profileview.List(pd, users, query))
	}
}

func (h *Handler) ProfileShow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user, err := repository.GetUser(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	recentInfos, _ := repository.ListChannelInfosByUser(h.DB, user.ID, 20)
	channels, _ := repository.ListChannelsByUser(h.DB, user.ID)

	pd := h.pageData(r, w)
	h.renderTempl(w, r, profileview.Show(pd, user, channels, recentInfos))
}

func (h *Handler) ProfileEdit(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	pd := h.pageData(r, w)
	h.renderTempl(w, r, profileview.Edit(pd, user))
}

func (h *Handler) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	name := r.FormValue("name")
	bio := r.FormValue("bio")

	if name == "" {
		h.flash(w, r, "名前を入力してください")
		http.Redirect(w, r, "/profile/edit", http.StatusFound)
		return
	}
	if len(bio) > 160 {
		h.flash(w, r, "自己紹介は160文字以内で入力してください")
		http.Redirect(w, r, "/profile/edit", http.StatusFound)
		return
	}

	repository.UpdateUser(h.DB, user.ID, name, user.Image, bio)
	h.flash(w, r, "プロフィールを更新しました")
	http.Redirect(w, r, "/profile/edit", http.StatusFound)
}
