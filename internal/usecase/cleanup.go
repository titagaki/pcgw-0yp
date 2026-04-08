package usecase

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/titagaki/pcgw-0yp/internal/domain"
	"github.com/titagaki/pcgw-0yp/internal/peercast"
	"github.com/titagaki/pcgw-0yp/internal/repository"
)

const InactiveTimeout = 15 * time.Minute

// RunCleanup checks all active channels and cleans up inactive ones.
// newClient creates a PeerCast client for a given servent.
func RunCleanup(db *sql.DB, log *slog.Logger, newClient func(*domain.Servent) *peercast.Client) {
	channels, err := repository.ListChannels(db)
	if err != nil {
		return
	}

	for _, ch := range channels {
		servent, err := repository.GetServent(db, ch.ServentID)
		if err != nil {
			// Servent deleted, remove channel
			repository.DeleteChannel(db, ch.ID)
			continue
		}

		client := newClient(servent)
		status, err := client.GetChannelStatus(ch.GnuID)
		if err != nil {
			// PeerCast unavailable, skip
			continue
		}

		if status.IsReceiving {
			repository.UpdateChannelLastActive(db, ch.ID)
		} else if ch.LastActiveAt.Valid && time.Since(ch.LastActiveAt.Time) > InactiveTimeout {
			log.Info("stopping inactive channel", "id", ch.ID, "gnu_id", ch.GnuID)
			client.StopChannel(ch.GnuID)
			client.RevokeStreamKey(fmt.Sprintf("user_%d", ch.UserID))
			repository.DeleteChannel(db, ch.ID)
		} else if !ch.LastActiveAt.Valid && time.Since(ch.CreatedAt) > InactiveTimeout {
			log.Info("stopping never-active channel", "id", ch.ID, "gnu_id", ch.GnuID)
			client.StopChannel(ch.GnuID)
			client.RevokeStreamKey(fmt.Sprintf("user_%d", ch.UserID))
			repository.DeleteChannel(db, ch.ID)
		}
	}
}
