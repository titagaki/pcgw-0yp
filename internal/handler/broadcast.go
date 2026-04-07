package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/model"
	"github.com/titagaki/pcgw-0yp/internal/peercast"
	channelview "github.com/titagaki/pcgw-0yp/internal/view/channel"
)

func (h *Handler) CreatePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	var tmplInfo *model.ChannelInfo
	if tmplID := r.URL.Query().Get("template"); tmplID != "" {
		if id, err := strconv.ParseInt(tmplID, 10, 64); err == nil {
			tmplInfo, _ = model.GetChannelInfo(h.DB, id)
		}
	}
	if tmplInfo == nil {
		tmplInfo, _ = model.GetLatestChannelInfoByUser(h.DB, user.ID)
	}

	servents, _ := model.ListEnabledServents(h.DB)
	sources, _ := model.ListSourcesByUser(h.DB, user.ID)

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

func generateStreamKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "sk_" + hex.EncodeToString(b)
}

func (h *Handler) Broadcast(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("channel_name")
	genre := r.FormValue("genre")
	desc := r.FormValue("desc")
	comment := r.FormValue("comment")
	contactURL := r.FormValue("url")
	ypName := r.FormValue("yp")
	sourceName := r.FormValue("source_name")

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

	var servent *model.Servent
	if serventID := r.FormValue("servent_id"); serventID != "" {
		if id, err := strconv.ParseInt(serventID, 10, 64); err == nil {
			servent, _ = model.GetServent(h.DB, id)
		}
	}
	if servent == nil {
		var err error
		servent, err = model.RequestServentWithVacancy(h.DB)
		if err != nil {
			h.flash(w, r, "利用可能なサーバーがありません")
			http.Redirect(w, r, "/create", http.StatusFound)
			return
		}
	}

	client := h.peercastClient(servent)

	streamKey := generateStreamKey()
	accountName := fmt.Sprintf("user_%d", user.ID)
	if err := client.IssueStreamKey(accountName, streamKey); err != nil {
		h.Log.Error("issueStreamKey failed", "error", err)
		h.flash(w, r, "ストリームキーの発行に失敗しました: "+err.Error())
		http.Redirect(w, r, "/create", http.StatusFound)
		return
	}

	bitrate, _ := strconv.Atoi(r.FormValue("bitrate"))

	breq := &peercast.BroadcastRequest{
		StreamKey: streamKey,
		Info: peercast.ChannelInfo{
			Name:    name,
			Genre:   genre,
			Desc:    desc,
			Comment: comment,
			URL:     contactURL,
			Bitrate: bitrate,
		},
		Track: peercast.TrackInfo{
			Creator: user.Name,
		},
	}

	result, err := client.BroadcastChannel(breq)
	if err != nil {
		h.Log.Error("broadcastChannel failed", "error", err)
		client.RevokeStreamKey(accountName)
		h.flash(w, r, "配信の開始に失敗しました: "+err.Error())
		http.Redirect(w, r, "/create", http.StatusFound)
		return
	}

	ch, err := model.CreateChannel(h.DB, result.ChannelID, user.ID, servent.ID, streamKey)
	if err != nil {
		h.Log.Error("create channel record failed", "error", err)
		client.StopChannel(result.ChannelID)
		client.RevokeStreamKey(accountName)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = model.CreateChannelInfo(h.DB, user.ID,
		name, genre, desc, comment, contactURL,
		"FLV", ypName,
		sql.NullInt64{Int64: ch.ID, Valid: true},
		sql.NullInt64{Int64: servent.ID, Valid: true},
		sourceName,
	)
	if err != nil {
		h.Log.Error("create channel_info failed", "error", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/channels/%d", ch.ID), http.StatusFound)
}
