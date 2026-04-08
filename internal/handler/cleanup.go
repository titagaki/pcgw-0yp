package handler

import (
	"context"
	"time"

	"github.com/titagaki/pcgw-0yp/internal/usecase"
)

const cleanupInterval = 30 * time.Second

// StartCleanup runs periodic channel cleanup in a goroutine.
func (h *Handler) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				usecase.RunCleanup(h.DB, h.Log, h.peercastClient)
			}
		}
	}()
}
