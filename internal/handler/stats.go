package handler

import (
	"net/http"

	"github.com/titagaki/pcgw-0yp/internal/repository"
	"github.com/titagaki/pcgw-0yp/internal/view/page"
)

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	servents, _ := repository.ListEnabledServents(h.DB)

	var statuses []page.ServentStatus
	for _, s := range servents {
		client := h.peercastClient(s)
		channels, err := client.GetChannels()
		statuses = append(statuses, page.ServentStatus{
			Servent:  s,
			Channels: channels,
			Error:    err,
		})
	}

	pd := h.pageData(r, w)
	h.renderTempl(w, r, page.Stats(pd, statuses))
}
