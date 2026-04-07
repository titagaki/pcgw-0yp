package handler

import (
	"fmt"
	"time"

	"github.com/titagaki/pcgw-0yp/internal/model"
)

const cleanupInterval = 30 * time.Second
const inactiveTimeout = 15 * time.Minute

// StartCleanup runs periodic channel cleanup in a goroutine.
func (h *Handler) StartCleanup() {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			h.runCleanup()
		}
	}()
}

func (h *Handler) runCleanup() {
	channels, err := model.ListChannels(h.DB)
	if err != nil {
		return
	}

	for _, ch := range channels {
		servent, err := model.GetServent(h.DB, ch.ServentID)
		if err != nil {
			// Servent deleted, remove channel
			model.DeleteChannel(h.DB, ch.ID)
			continue
		}

		client := h.peercastClient(servent)
		status, err := client.GetChannelStatus(ch.GnuID)
		if err != nil {
			// PeerCast unavailable, skip
			continue
		}

		if status.IsReceiving {
			model.UpdateChannelLastActive(h.DB, ch.ID)
		} else if ch.LastActiveAt.Valid && time.Since(ch.LastActiveAt.Time) > inactiveTimeout {
			// Inactive for too long, stop channel
			h.Log.Info("stopping inactive channel", "id", ch.ID, "gnu_id", ch.GnuID)
			client.StopChannel(ch.GnuID)
			client.RevokeStreamKey(fmt.Sprintf("user_%d", ch.UserID))
			model.DeleteChannel(h.DB, ch.ID)
		} else if !ch.LastActiveAt.Valid && time.Since(ch.CreatedAt) > inactiveTimeout {
			// Never received data, stop
			h.Log.Info("stopping never-active channel", "id", ch.ID, "gnu_id", ch.GnuID)
			client.StopChannel(ch.GnuID)
			client.RevokeStreamKey(fmt.Sprintf("user_%d", ch.UserID))
			model.DeleteChannel(h.DB, ch.ID)
		}
	}
}
