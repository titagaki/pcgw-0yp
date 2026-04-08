package handler

import (
	"encoding/json"
	"net/http"

	"github.com/titagaki/pcgw-0yp/internal/model"
)

func (h *Handler) APIChannelStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	name := r.URL.Query().Get("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name parameter required"})
		return
	}

	// servent ごとにチャンネル一覧を取得し、名前で探す
	servents, _ := model.ListEnabledServents(h.DB)
	for _, servent := range servents {
		client := h.peercastClient(servent)
		entries, err := client.GetChannels()
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.Info.Name != name {
				continue
			}
			result := map[string]interface{}{
				"name":           entry.Info.Name,
				"genre":          entry.Info.Genre,
				"desc":           entry.Info.Desc,
				"comment":        entry.Info.Comment,
				"url":            entry.Info.URL,
				"bitrate":        entry.Info.Bitrate,
				"contentType":    entry.Info.ContentType,
				"listeners":      entry.Status.TotalDirects,
				"relays":         entry.Status.TotalRelays,
				"status":         entry.Status.Status,
				"uptime":         entry.Status.Uptime,
				"isBroadcasting": entry.Status.IsBroadcasting,
			}
			if ch, err := model.GetChannelByGnuID(h.DB, entry.ChannelID); err == nil {
				if ci, err := model.GetChannelInfoByChannelID(h.DB, ch.ID); err == nil {
					result["streamType"] = ci.StreamType
				}
			}
			json.NewEncoder(w).Encode(result)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "channel not found"})
}
