package peercast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

type Client struct {
	url      string
	authID   string
	passwd   string
	client   *http.Client
	idSeq    atomic.Int64
}

func NewClient(hostname string, port int, authID, passwd string) *Client {
	return &Client{
		url:    fmt.Sprintf("http://%s:%d/api/1", hostname, port),
		authID: authID,
		passwd: passwd,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int64       `json:"id"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
	ID      int64           `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

type Unavailable struct {
	Host    string
	Port    int
	Message string
}

func (e *Unavailable) Error() string {
	return fmt.Sprintf("PeerCast unavailable at %s:%d: %s", e.Host, e.Port, e.Message)
}

func (c *Client) call(method string, params interface{}) (json.RawMessage, error) {
	id := c.idSeq.Add(1)
	req := rpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Requested-With", "XMLHttpRequest")
	if c.authID != "" {
		httpReq.SetBasicAuth(c.authID, c.passwd)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, &Unavailable{Message: err.Error()}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("invalid JSON-RPC response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}
	return rpcResp.Result, nil
}

// VersionInfo is the result of getVersionInfo.
type VersionInfo struct {
	AgentName string `json:"agentName"`
}

func (c *Client) GetVersionInfo() (*VersionInfo, error) {
	result, err := c.call("getVersionInfo", []interface{}{})
	if err != nil {
		return nil, err
	}
	var v VersionInfo
	return &v, json.Unmarshal(result, &v)
}

// Settings is the result of getSettings.
type Settings struct {
	ServerPort int `json:"serverPort"`
	RTMPPort   int `json:"rtmpPort"`
}

func (c *Client) GetSettings() (*Settings, error) {
	result, err := c.call("getSettings", []interface{}{})
	if err != nil {
		return nil, err
	}
	var s Settings
	return &s, json.Unmarshal(result, &s)
}

// YellowPage is a YP entry from getYellowPages.
type YellowPage struct {
	YellowPageID int    `json:"yellowPageId"`
	Name         string `json:"name"`
	URI          string `json:"uri"`
	AnnounceURI  string `json:"announceUri"`
	ChannelCount int    `json:"channelCount"`
}

func (c *Client) GetYellowPages() ([]YellowPage, error) {
	result, err := c.call("getYellowPages", []interface{}{})
	if err != nil {
		return nil, err
	}
	var yps []YellowPage
	return yps, json.Unmarshal(result, &yps)
}

// ChannelEntry is a channel from getChannels.
type ChannelEntry struct {
	ChannelID string        `json:"channelId"`
	Status    ChannelStatus `json:"status"`
	Info      ChannelInfo   `json:"info"`
	Track     TrackInfo     `json:"track"`
}

type ChannelInfo struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Genre       string `json:"genre"`
	Desc        string `json:"desc"`
	Comment     string `json:"comment"`
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"contentType"`
	MIMEType    string `json:"mimeType"`
}

type TrackInfo struct {
	Title   string `json:"title"`
	Genre   string `json:"genre"`
	Album   string `json:"album"`
	Creator string `json:"creator"`
	URL     string `json:"url"`
}

type ChannelStatus struct {
	Status         string `json:"status"`
	Source         string `json:"source"`
	Uptime         int    `json:"uptime"`
	LocalRelays    int    `json:"localRelays"`
	LocalDirects   int    `json:"localDirects"`
	TotalRelays    int    `json:"totalRelays"`
	TotalDirects   int    `json:"totalDirects"`
	IsBroadcasting bool   `json:"isBroadcasting"`
	IsRelayFull    bool   `json:"isRelayFull"`
	IsDirectFull   bool   `json:"isDirectFull"`
	IsReceiving    bool   `json:"isReceiving"`
}

func (c *Client) GetChannels() ([]ChannelEntry, error) {
	result, err := c.call("getChannels", []interface{}{})
	if err != nil {
		return nil, err
	}
	var chs []ChannelEntry
	return chs, json.Unmarshal(result, &chs)
}

func (c *Client) GetChannelInfo(channelID string) (*struct {
	Info  ChannelInfo `json:"info"`
	Track TrackInfo   `json:"track"`
}, error) {
	result, err := c.call("getChannelInfo", []interface{}{channelID})
	if err != nil {
		return nil, err
	}
	var info struct {
		Info  ChannelInfo `json:"info"`
		Track TrackInfo   `json:"track"`
	}
	return &info, json.Unmarshal(result, &info)
}

func (c *Client) GetChannelStatus(channelID string) (*ChannelStatus, error) {
	result, err := c.call("getChannelStatus", []interface{}{channelID})
	if err != nil {
		return nil, err
	}
	var s ChannelStatus
	return &s, json.Unmarshal(result, &s)
}

// Connection is a connection entry from getChannelConnections.
type Connection struct {
	ConnectionID int    `json:"connectionId"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	SendRate     int64  `json:"sendRate"`
	RecvRate     int64  `json:"recvRate"`
	ProtocolName string `json:"protocolName"`
	RemoteEndPoint string `json:"remoteEndPoint"`
}

func (c *Client) GetChannelConnections(channelID string) ([]Connection, error) {
	result, err := c.call("getChannelConnections", []interface{}{channelID})
	if err != nil {
		return nil, err
	}
	var conns []Connection
	return conns, json.Unmarshal(result, &conns)
}

// RelayTreeNode is a node in the relay tree from getChannelRelayTree.
type RelayTreeNode struct {
	SessionID     string          `json:"sessionId"`
	Address       string          `json:"address"`
	Port          int             `json:"port"`
	IsFirewalled  bool            `json:"isFirewalled"`
	LocalRelays   int             `json:"localRelays"`
	LocalDirects  int             `json:"localDirects"`
	IsTracker     bool            `json:"isTracker"`
	IsRelayFull   bool            `json:"isRelayFull"`
	IsDirectFull  bool            `json:"isDirectFull"`
	IsReceiving   bool            `json:"isReceiving"`
	IsControlFull bool            `json:"isControlFull"`
	Version       int             `json:"version"`
	VersionString string          `json:"versionString"`
	Children      []RelayTreeNode `json:"children"`
}

func (c *Client) GetChannelRelayTree(channelID string) ([]RelayTreeNode, error) {
	result, err := c.call("getChannelRelayTree", []interface{}{channelID})
	if err != nil {
		return nil, err
	}
	var nodes []RelayTreeNode
	return nodes, json.Unmarshal(result, &nodes)
}

func (c *Client) SetChannelInfo(channelID string, info map[string]interface{}, track map[string]interface{}) error {
	_, err := c.call("setChannelInfo", []interface{}{channelID, info, track})
	return err
}

func (c *Client) StopChannel(channelID string) error {
	_, err := c.call("stopChannel", []interface{}{channelID})
	return err
}

func (c *Client) BumpChannel(channelID string) error {
	_, err := c.call("bumpChannel", []interface{}{channelID})
	return err
}

func (c *Client) StopChannelConnection(channelID string, connectionID int) (bool, error) {
	result, err := c.call("stopChannelConnection", []interface{}{channelID, connectionID})
	if err != nil {
		return false, err
	}
	var ok bool
	return ok, json.Unmarshal(result, &ok)
}

// Stream key management

func (c *Client) IssueStreamKey(accountName, streamKey string) error {
	_, err := c.call("issueStreamKey", []interface{}{accountName, streamKey})
	return err
}

func (c *Client) RevokeStreamKey(accountName string) error {
	_, err := c.call("revokeStreamKey", []interface{}{accountName})
	return err
}

type StreamKeyEntry struct {
	AccountName string `json:"accountName"`
	StreamKey   string `json:"streamKey"`
}

func (c *Client) ListStreamKeys() ([]StreamKeyEntry, error) {
	result, err := c.call("listStreamKeys", []interface{}{})
	if err != nil {
		return nil, err
	}
	var keys []StreamKeyEntry
	return keys, json.Unmarshal(result, &keys)
}

// BroadcastChannel creates a new broadcast channel.
type BroadcastRequest struct {
	StreamKey string      `json:"streamKey"`
	Info      ChannelInfo `json:"info"`
	Track     TrackInfo   `json:"track"`
}

type BroadcastResult struct {
	ChannelID string `json:"channelId"`
}

func (c *Client) BroadcastChannel(req *BroadcastRequest) (*BroadcastResult, error) {
	result, err := c.call("broadcastChannel", []interface{}{req})
	if err != nil {
		return nil, err
	}
	var br BroadcastResult
	return &br, json.Unmarshal(result, &br)
}
