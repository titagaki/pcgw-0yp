package usecase

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/titagaki/pcgw-0yp/internal/domain"
	"github.com/titagaki/pcgw-0yp/internal/peercast"
	"github.com/titagaki/pcgw-0yp/internal/repository"
)

type BroadcastParams struct {
	Name       string
	Genre      string
	Desc       string
	Comment    string
	URL        string
	YP         string
	SourceName string
	Bitrate    int
}

type BroadcastResult struct {
	Channel *domain.Channel
	PushURL string
}

func generateStreamKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "sk_" + hex.EncodeToString(b), nil
}

// StartBroadcast orchestrates the full broadcast flow:
// issue stream key -> broadcast channel on PeerCast -> create DB records.
// On failure, it rolls back already-completed steps.
func StartBroadcast(db *sql.DB, log *slog.Logger, client *peercast.Client, user *domain.User, servent *domain.Servent, params BroadcastParams) (*BroadcastResult, error) {
	streamKey, err := generateStreamKey()
	if err != nil {
		return nil, fmt.Errorf("generate stream key: %w", err)
	}

	accountName := fmt.Sprintf("user_%d", user.ID)
	if err := client.IssueStreamKey(accountName, streamKey); err != nil {
		return nil, fmt.Errorf("issue stream key: %w", err)
	}

	breq := &peercast.BroadcastRequest{
		StreamKey: streamKey,
		Info: peercast.ChannelInfo{
			Name:    params.Name,
			Genre:   params.Genre,
			Desc:    params.Desc,
			Comment: params.Comment,
			URL:     params.URL,
			Bitrate: params.Bitrate,
		},
		Track: peercast.TrackInfo{
			Creator: user.Name,
		},
	}

	pcResult, err := client.BroadcastChannel(breq)
	if err != nil {
		log.Error("broadcastChannel failed", "error", err)
		client.RevokeStreamKey(accountName)
		return nil, fmt.Errorf("broadcast channel: %w", err)
	}

	ch, err := repository.CreateChannel(db, pcResult.ChannelID, user.ID, servent.ID, streamKey)
	if err != nil {
		log.Error("create channel record failed", "error", err)
		client.StopChannel(pcResult.ChannelID)
		client.RevokeStreamKey(accountName)
		return nil, fmt.Errorf("create channel: %w", err)
	}

	_, err = repository.CreateChannelInfo(db, user.ID,
		params.Name, params.Genre, params.Desc, params.Comment, params.URL,
		"FLV", params.YP,
		sql.NullInt64{Int64: ch.ID, Valid: true},
		sql.NullInt64{Int64: servent.ID, Valid: true},
		params.SourceName,
	)
	if err != nil {
		log.Error("create channel_info failed", "error", err)
	}

	return &BroadcastResult{Channel: ch}, nil
}

// StopChannel stops a broadcast: stop on PeerCast, revoke stream key, delete DB record.
func StopChannel(db *sql.DB, client *peercast.Client, ch *domain.Channel) {
	client.StopChannel(ch.GnuID)
	client.RevokeStreamKey(fmt.Sprintf("user_%d", ch.UserID))
	repository.DeleteChannel(db, ch.ID)
}

// UpdateChannelOnPeerCast updates channel info on PeerCast and in DB.
func UpdateChannelOnPeerCast(db *sql.DB, log *slog.Logger, client *peercast.Client, ch *domain.Channel, userName string, name, genre, desc, comment, contactURL string) {
	info := map[string]interface{}{
		"name":    name,
		"genre":   genre,
		"desc":    desc,
		"comment": comment,
		"url":     contactURL,
	}
	track := map[string]interface{}{
		"creator": userName,
	}
	if err := client.SetChannelInfo(ch.GnuID, info, track); err != nil {
		log.Error("setChannelInfo failed", "error", err)
	}

	if ci, err := repository.GetChannelInfoByChannelID(db, ch.ID); err == nil {
		repository.UpdateChannelInfo(db, ci.ID, name, genre, desc, comment, contactURL)
	}
}
