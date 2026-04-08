package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/repository"
	"github.com/titagaki/pcgw-0yp/internal/usecase"
	channelview "github.com/titagaki/pcgw-0yp/internal/view/channel"
)

func (h *Handler) ChannelList(w http.ResponseWriter, r *http.Request) {
	channels, _ := repository.ListChannels(h.DB)
	pd := h.pageData(r, w)
	h.renderTempl(w, r, channelview.List(pd, channels))
}

func (h *Handler) ChannelShow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ch, err := repository.GetChannel(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.CurrentUser(r)
	if user == nil || (user.ID != ch.UserID && !user.Admin) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	servent, _ := repository.GetServent(h.DB, ch.ServentID)
	channelInfo, _ := repository.GetChannelInfoByChannelID(h.DB, ch.ID)

	var status *channelview.StatusData
	if servent != nil {
		client := h.peercastClient(servent)
		if cs, err := client.GetChannelStatus(ch.GnuID); err == nil {
			conns, _ := client.GetChannelConnections(ch.GnuID)
			status = &channelview.StatusData{
				Status:      cs,
				Connections: conns,
			}
			if cs.IsReceiving {
				repository.UpdateChannelLastActive(h.DB, ch.ID)
			}
		}
	}

	var pushURL string
	if servent != nil {
		pushURL = buildRTMPPushURL(servent.Hostname, 1935, ch.StreamKey)
	}

	pd := h.pageData(r, w)
	h.renderTempl(w, r, channelview.Show(pd, channelview.ShowData{
		Channel:     ch,
		ChannelInfo: channelInfo,
		Servent:     servent,
		Status:      status,
		PushURL:     pushURL,
	}))
}

func (h *Handler) ChannelUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ch, err := repository.GetChannel(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.CurrentUser(r)
	if user == nil || (user.ID != ch.UserID && !user.Admin) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("channel_name")
	genre := r.FormValue("genre")
	desc := r.FormValue("desc")
	comment := r.FormValue("comment")
	contactURL := r.FormValue("url")

	servent, _ := repository.GetServent(h.DB, ch.ServentID)
	if servent != nil {
		client := h.peercastClient(servent)
		usecase.UpdateChannelOnPeerCast(h.DB, h.Log, client, ch, user.Name, name, genre, desc, comment, contactURL)
	}

	h.flash(w, r, "チャンネル情報を更新しました")
	http.Redirect(w, r, fmt.Sprintf("/channels/%d", ch.ID), http.StatusFound)
}

func (h *Handler) ChannelEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ch, err := repository.GetChannel(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.CurrentUser(r)
	if user == nil || (user.ID != ch.UserID && !user.Admin) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	channelInfo, _ := repository.GetChannelInfoByChannelID(h.DB, ch.ID)
	pd := h.pageData(r, w)
	h.renderTempl(w, r, channelview.Edit(pd, ch, channelInfo))
}

func (h *Handler) ChannelStop(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ch, err := repository.GetChannel(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.CurrentUser(r)
	if user == nil || (user.ID != ch.UserID && !user.Admin) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	servent, _ := repository.GetServent(h.DB, ch.ServentID)
	if servent != nil {
		client := h.peercastClient(servent)
		usecase.StopChannel(h.DB, client, ch)
	} else {
		repository.DeleteChannel(h.DB, ch.ID)
	}

	h.flash(w, r, "配信を停止しました")
	http.Redirect(w, r, "/home", http.StatusFound)
}

func (h *Handler) ChannelRelayTree(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ch, err := repository.GetChannel(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	servent, _ := repository.GetServent(h.DB, ch.ServentID)
	if servent == nil {
		http.NotFound(w, r)
		return
	}

	client := h.peercastClient(servent)
	tree, err := client.GetChannelRelayTree(ch.GnuID)
	if err != nil {
		h.Log.Error("getChannelRelayTree failed", "error", err)
		tree = nil
	}

	pd := h.pageData(r, w)
	h.renderTempl(w, r, channelview.RelayTree(pd, ch, tree))
}

func (h *Handler) ChannelConnections(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ch, err := repository.GetChannel(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	servent, _ := repository.GetServent(h.DB, ch.ServentID)
	if servent == nil {
		http.NotFound(w, r)
		return
	}

	client := h.peercastClient(servent)
	conns, err := client.GetChannelConnections(ch.GnuID)
	if err != nil {
		h.Log.Error("getChannelConnections failed", "error", err)
	}

	pd := h.pageData(r, w)
	h.renderTempl(w, r, channelview.Connections(pd, ch, conns))
}

func (h *Handler) ChannelDisconnect(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ch, err := repository.GetChannel(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.CurrentUser(r)
	if user == nil || (user.ID != ch.UserID && !user.Admin) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	connID, err := strconv.Atoi(chi.URLParam(r, "connID"))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	servent, _ := repository.GetServent(h.DB, ch.ServentID)
	if servent != nil {
		client := h.peercastClient(servent)
		client.StopChannelConnection(ch.GnuID, connID)
	}

	http.Redirect(w, r, fmt.Sprintf("/channels/%d", ch.ID), http.StatusFound)
}

func (h *Handler) ChannelStatusJSON(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ch, err := repository.GetChannel(h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	servent, _ := repository.GetServent(h.DB, ch.ServentID)
	if servent == nil {
		http.NotFound(w, r)
		return
	}

	client := h.peercastClient(servent)
	cs, err := client.GetChannelStatus(ch.GnuID)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if cs.IsReceiving {
		repository.UpdateChannelLastActive(h.DB, ch.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cs)
}
