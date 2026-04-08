package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/titagaki/pcgw-0yp/internal/domain"
	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/peercast"
	"github.com/titagaki/pcgw-0yp/internal/repository"
	"github.com/titagaki/pcgw-0yp/internal/usecase"
	channelview "github.com/titagaki/pcgw-0yp/internal/view/channel"
)

func (h *Handler) CreatePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	var tmplInfo *domain.ChannelInfo
	if tmplID := r.URL.Query().Get("template"); tmplID != "" {
		if id, err := strconv.ParseInt(tmplID, 10, 64); err == nil {
			tmplInfo, _ = repository.GetChannelInfo(h.DB, id)
		}
	}
	if tmplInfo == nil {
		tmplInfo, _ = repository.GetLatestChannelInfoByUser(h.DB, user.ID)
	}

	servents, _ := repository.ListEnabledServents(h.DB)
	sources, _ := repository.ListSourcesByUser(h.DB, user.ID)

	var yps []peercast.YellowPage
	if len(servents) > 0 {
		client := h.peercastClient(servents[0])
		yps, _ = client.GetYellowPages()
	}

	pd := h.pageData(r, w)
	h.renderTempl(w, r, channelview.Create(pd, channelview.CreateData{
		Template: tmplInfo,
		Servents: servents,
		Sources:  sources,
		YPs:      yps,
	}))
}

func (h *Handler) Broadcast(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("channel_name")
	if name == "" {
		h.flash(w, r, "チャンネル名を入力してください")
		http.Redirect(w, r, "/create", http.StatusFound)
		return
	}
	if len(name) > 255 {
		h.flash(w, r, "チャンネル名が長すぎます")
		http.Redirect(w, r, "/create", http.StatusFound)
		return
	}

	var servent *domain.Servent
	if serventID := r.FormValue("servent_id"); serventID != "" {
		if id, err := strconv.ParseInt(serventID, 10, 64); err == nil {
			servent, _ = repository.GetServent(h.DB, id)
		}
	}
	if servent == nil {
		var err error
		servent, err = repository.RequestServentWithVacancy(h.DB)
		if err != nil {
			h.flash(w, r, "利用可能なサーバーがありません")
			http.Redirect(w, r, "/create", http.StatusFound)
			return
		}
	}

	client := h.peercastClient(servent)
	bitrate, _ := strconv.Atoi(r.FormValue("bitrate"))

	result, err := usecase.StartBroadcast(h.DB, h.Log, client, user, servent, usecase.BroadcastParams{
		Name:       name,
		Genre:      r.FormValue("genre"),
		Desc:       r.FormValue("desc"),
		Comment:    r.FormValue("comment"),
		URL:        r.FormValue("url"),
		YP:         r.FormValue("yp"),
		SourceName: r.FormValue("source_name"),
		Bitrate:    bitrate,
	})
	if err != nil {
		h.flash(w, r, "配信の開始に失敗しました: "+err.Error())
		http.Redirect(w, r, "/create", http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/channels/%d", result.Channel.ID), http.StatusFound)
}
