package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/titagaki/pcgw-0yp/internal/model"
	"github.com/titagaki/pcgw-0yp/internal/peercast"
	adminview "github.com/titagaki/pcgw-0yp/internal/view/admin"
)

func (h *Handler) ServentIndex(w http.ResponseWriter, r *http.Request) {
	servents, _ := model.ListServents(h.DB)
	pd := h.pageData(r, w)
	h.renderTempl(w, r, adminview.ServentList(pd, servents))
}

func (h *Handler) ServentShow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	servent, err := model.GetServent(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var channels []peercast.ChannelEntry
	if servent.Enabled {
		client := h.peercastClient(servent)
		channels, _ = client.GetChannels()
	}

	pd := h.pageData(r, w)
	h.renderTempl(w, r, adminview.ServentShow(pd, servent, channels))
}

func (h *Handler) ServentCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	desc := r.FormValue("desc")
	hostname := r.FormValue("hostname")
	port, _ := strconv.Atoi(r.FormValue("port"))
	authID := r.FormValue("auth_id")
	passwd := r.FormValue("passwd")
	priority, _ := strconv.Atoi(r.FormValue("priority"))
	maxChannels, _ := strconv.Atoi(r.FormValue("max_channels"))
	enabled := r.FormValue("enabled") == "on"

	if name == "" || hostname == "" || port == 0 {
		h.flash(w, r, "名前、ホスト名、ポートは必須です")
		http.Redirect(w, r, "/servents", http.StatusFound)
		return
	}

	_, err := model.CreateServent(h.DB, name, desc, hostname, port, authID, passwd, priority, maxChannels, enabled)
	if err != nil {
		h.flash(w, r, "作成に失敗しました: "+err.Error())
		http.Redirect(w, r, "/servents", http.StatusFound)
		return
	}

	h.flash(w, r, "サーバントを追加しました")
	http.Redirect(w, r, "/servents", http.StatusFound)
}

func (h *Handler) ServentUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	desc := r.FormValue("desc")
	hostname := r.FormValue("hostname")
	port, _ := strconv.Atoi(r.FormValue("port"))
	authID := r.FormValue("auth_id")
	passwd := r.FormValue("passwd")
	priority, _ := strconv.Atoi(r.FormValue("priority"))
	maxChannels, _ := strconv.Atoi(r.FormValue("max_channels"))
	enabled := r.FormValue("enabled") == "on"

	if err := model.UpdateServent(h.DB, id, name, desc, hostname, port, authID, passwd, priority, maxChannels, enabled); err != nil {
		h.flash(w, r, "更新に失敗しました: "+err.Error())
	} else {
		h.flash(w, r, "サーバントを更新しました")
	}

	http.Redirect(w, r, fmt.Sprintf("/servents/%d", id), http.StatusFound)
}

func (h *Handler) ServentDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	model.DeleteServent(h.DB, id)
	h.flash(w, r, "サーバントを削除しました")
	http.Redirect(w, r, "/servents", http.StatusFound)
}

func (h *Handler) ServentRefresh(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	servent, err := model.GetServent(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	client := h.peercastClient(servent)
	versionInfo, err := client.GetVersionInfo()
	if err != nil {
		h.flash(w, r, "サーバーに接続できません: "+err.Error())
		http.Redirect(w, r, fmt.Sprintf("/servents/%d", id), http.StatusFound)
		return
	}

	yps, _ := client.GetYellowPages()
	var ypNames []string
	for _, yp := range yps {
		ypNames = append(ypNames, yp.Name)
	}

	model.UpdateServentAgent(h.DB, id, versionInfo.AgentName, strings.Join(ypNames, " "))
	h.flash(w, r, "サーバー情報を更新しました")
	http.Redirect(w, r, fmt.Sprintf("/servents/%d", id), http.StatusFound)
}
