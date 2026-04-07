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

	channels, _ := model.ListChannels(h.DB)
	for _, ch := range channels {
		servent, _ := model.GetServent(h.DB, ch.ServentID)
		if servent == nil {
			continue
		}
		client := h.peercastClient(servent)
		info, err := client.GetChannelInfo(ch.GnuID)
		if err != nil {
			continue
		}
		if info.Info.Name == name {
			status, err := client.GetChannelStatus(ch.GnuID)
			if err != nil {
				continue
			}
			ci, _ := model.GetChannelInfoByChannelID(h.DB, ch.ID)
			result := map[string]interface{}{
				"name":         info.Info.Name,
				"genre":        info.Info.Genre,
				"desc":         info.Info.Desc,
				"comment":      info.Info.Comment,
				"url":          info.Info.URL,
				"bitrate":      info.Info.Bitrate,
				"contentType":  info.Info.ContentType,
				"listeners":    status.TotalDirects,
				"relays":       status.TotalRelays,
				"status":       status.Status,
				"uptime":       status.Uptime,
				"isBroadcasting": status.IsBroadcasting,
			}
			if ci != nil {
				result["streamType"] = ci.StreamType
			}
			json.NewEncoder(w).Encode(result)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "channel not found"})
}
