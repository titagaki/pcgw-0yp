package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/titagaki/pcgw-0yp/internal/model"
)

func (h *Handler) AdminIndex(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "admin.html", nil)
}

func (h *Handler) UserList(w http.ResponseWriter, r *http.Request) {
	users, _ := model.ListUsers(h.DB)
	data := map[string]interface{}{
		"Users":   users,
		"Flashes": h.getFlashes(r, w),
	}
	h.render(w, r, "users.html", data)
}

func (h *Handler) UserShow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := model.GetUser(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data := map[string]interface{}{
		"TargetUser": user,
	}
	h.render(w, r, "user.html", data)
}

func (h *Handler) UserEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := model.GetUser(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data := map[string]interface{}{
		"TargetUser": user,
		"Flashes":    h.getFlashes(r, w),
	}
	h.render(w, r, "user_edit.html", data)
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

	user, _ := model.GetUser(h.DB, id)
	if user == nil {
		http.NotFound(w, r)
		return
	}

	model.UpdateUser(h.DB, id, name, image, user.Bio)
	model.UpdateUserAdmin(h.DB, id, admin)
	model.UpdateUserSuspended(h.DB, id, suspended)

	h.flash(w, r, "ユーザーを更新しました")
	http.Redirect(w, r, fmt.Sprintf("/users/%d/edit", id), http.StatusFound)
}

func (h *Handler) UserDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	model.DeleteUser(h.DB, id)
	h.flash(w, r, "ユーザーを削除しました")
	http.Redirect(w, r, "/users", http.StatusFound)
}
