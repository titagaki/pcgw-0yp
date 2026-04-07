package handler

import (
	"net/http"

	"github.com/titagaki/pcgw-0yp/internal/model"
	"github.com/titagaki/pcgw-0yp/internal/peercast"
)

type serventStatus struct {
	Servent  *model.Servent
	Channels []peercast.ChannelEntry
	Error    error
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	servents, _ := model.ListEnabledServents(h.DB)

	var statuses []serventStatus
	for _, s := range servents {
		client := h.peercastClient(s)
		channels, err := client.GetChannels()
		statuses = append(statuses, serventStatus{
			Servent:  s,
			Channels: channels,
			Error:    err,
		})
	}

	data := map[string]interface{}{
		"ServentStatuses": statuses,
	}
	h.render(w, r, "stats.html", data)
}
