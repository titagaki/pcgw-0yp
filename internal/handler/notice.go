package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/model"
)

func (h *Handler) NoticeIndex(w http.ResponseWriter, r *http.Request) {
	notices, _ := model.ListNotices(h.DB)

	if user := middleware.CurrentUser(r); user != nil {
		model.UpdateUserNoticeChecked(h.DB, user.ID)
	}

	data := map[string]interface{}{
		"Notices": notices,
		"Flashes": h.getFlashes(r, w),
	}
	h.render(w, r, "notices.html", data)
}

func (h *Handler) NoticeShow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	notice, err := model.GetNotice(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data := map[string]interface{}{
		"Notice": notice,
	}
	h.render(w, r, "notice.html", data)
}

func (h *Handler) NoticeNew(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "notice_form.html", nil)
}

func (h *Handler) NoticeCreate(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")

	if title == "" || body == "" {
		h.flash(w, r, "タイトルと本文は必須です")
		http.Redirect(w, r, "/notices/new", http.StatusFound)
		return
	}

	notice, err := model.CreateNotice(h.DB, title, body)
	if err != nil {
		h.flash(w, r, "作成に失敗しました")
		http.Redirect(w, r, "/notices/new", http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/notices/%d", notice.ID), http.StatusFound)
}

func (h *Handler) NoticeEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	notice, err := model.GetNotice(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data := map[string]interface{}{
		"Notice": notice,
	}
	h.render(w, r, "notice_edit.html", data)
}

func (h *Handler) NoticeUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	model.UpdateNotice(h.DB, id, title, body)
	h.flash(w, r, "お知らせを更新しました")
	http.Redirect(w, r, fmt.Sprintf("/notices/%d", id), http.StatusFound)
}

func (h *Handler) NoticeDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	model.DeleteNotice(h.DB, id)
	h.flash(w, r, "お知らせを削除しました")
	http.Redirect(w, r, "/notices", http.StatusFound)
}
