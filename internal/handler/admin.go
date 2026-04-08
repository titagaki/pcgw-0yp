package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/titagaki/pcgw-0yp/internal/repository"
	adminview "github.com/titagaki/pcgw-0yp/internal/view/admin"
)

func (h *Handler) AdminIndex(w http.ResponseWriter, r *http.Request) {
	pd := h.pageData(r, w)
	h.renderTempl(w, r, adminview.Index(pd))
}

func (h *Handler) UserList(w http.ResponseWriter, r *http.Request) {
	users, _ := repository.ListUsers(h.DB)
	pd := h.pageData(r, w)
	h.renderTempl(w, r, adminview.UserList(pd, users))
}

func (h *Handler) UserShow(w http.ResponseWriter, r *http.Request) {
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
	pd := h.pageData(r, w)
	h.renderTempl(w, r, adminview.UserShow(pd, user))
}

func (h *Handler) UserEdit(w http.ResponseWriter, r *http.Request) {
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
	pd := h.pageData(r, w)
	h.renderTempl(w, r, adminview.UserEdit(pd, user))
}

func (h *Handler) UserUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	name := r.FormValue("name")
	image := r.FormValue("image")
	admin := r.FormValue("admin") == "on"
	suspended := r.FormValue("suspended") == "on"

	user, _ := repository.GetUser(h.DB, id)
	if user == nil {
		http.NotFound(w, r)
		return
	}

	repository.UpdateUser(h.DB, id, name, image, user.Bio)
	repository.UpdateUserAdmin(h.DB, id, admin)
	repository.UpdateUserSuspended(h.DB, id, suspended)

	h.flash(w, r, "ユーザーを更新しました")
	http.Redirect(w, r, fmt.Sprintf("/users/%d/edit", id), http.StatusFound)
}

func (h *Handler) UserDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	repository.DeleteUser(h.DB, id)
	h.flash(w, r, "ユーザーを削除しました")
	http.Redirect(w, r, "/users", http.StatusFound)
}
