package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/model"
	"github.com/titagaki/pcgw-0yp/internal/view/page"
	"golang.org/x/oauth2"
)

func (h *Handler) twitterOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.Config.Twitter.ClientID,
		ClientSecret: h.Config.Twitter.ClientSecret,
		RedirectURL:  h.Config.Twitter.RedirectURL,
		Scopes:       []string{"tweet.read", "users.read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://twitter.com/i/oauth2/authorize",
			TokenURL: "https://api.twitter.com/2/oauth2/token",
		},
	}
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if middleware.IsLoggedIn(r) {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}
	pd := h.pageData(r, w)
	backref := r.URL.Query().Get("backref")
	h.renderTempl(w, r, page.Login(pd, backref))
}

func (h *Handler) TwitterLogin(w http.ResponseWriter, r *http.Request) {
	cfg := h.twitterOAuth2Config()
	session := middleware.GetSession(r)

	// Generate PKCE code verifier
	verifier := make([]byte, 32)
	rand.Read(verifier)
	codeVerifier := hex.EncodeToString(verifier)
	session.Values["code_verifier"] = codeVerifier

	// Generate state
	stateBytes := make([]byte, 16)
	rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)
	session.Values["oauth_state"] = state

	if backref := r.URL.Query().Get("backref"); backref != "" {
		session.Values["oauth_backref"] = backref
	}
	session.Save(r, w)

	authURL := cfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeVerifier),
		oauth2.SetAuthURLParam("code_challenge_method", "plain"),
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Handler) TwitterCallback(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r)

	// Verify state
	expectedState, _ := session.Values["oauth_state"].(string)
	if r.URL.Query().Get("state") != expectedState {
		h.flash(w, r, "認証に失敗しました")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	cfg := h.twitterOAuth2Config()
	codeVerifier, _ := session.Values["code_verifier"].(string)

	token, err := cfg.Exchange(context.Background(), r.URL.Query().Get("code"),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		h.Log.Error("twitter oauth exchange failed", "error", err)
		h.flash(w, r, "認証に失敗しました")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Fetch user info
	client := cfg.Client(context.Background(), token)
	resp, err := client.Get("https://api.twitter.com/2/users/me?user.fields=profile_image_url")
	if err != nil {
		h.Log.Error("twitter user fetch failed", "error", err)
		h.flash(w, r, "ユーザー情報の取得に失敗しました")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var twitterUser struct {
		Data struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Username        string `json:"username"`
			ProfileImageURL string `json:"profile_image_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &twitterUser); err != nil || twitterUser.Data.ID == "" {
		h.Log.Error("twitter user parse failed", "error", err, "body", string(body))
		h.flash(w, r, "ユーザー情報の取得に失敗しました")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Get larger profile image
	imageURL := strings.Replace(twitterUser.Data.ProfileImageURL, "_normal", "_200x200", 1)

	// Find or create user
	user, err := model.GetUserByTwitterID(h.DB, twitterUser.Data.ID)
	if err == sql.ErrNoRows {
		// New user
		name := twitterUser.Data.Name
		if name == "" {
			name = twitterUser.Data.Username
		}
		user, err = model.CreateUser(h.DB, name, imageURL, twitterUser.Data.ID)
		if err != nil {
			h.Log.Error("user create failed", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		h.Log.Error("user lookup failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		// Update profile image
		model.UpdateUser(h.DB, user.ID, user.Name, imageURL, user.Bio)
	}

	// Set session
	session.Values["uid"] = user.ID
	delete(session.Values, "oauth_state")
	delete(session.Values, "code_verifier")
	backref, _ := session.Values["oauth_backref"].(string)
	delete(session.Values, "oauth_backref")
	session.Save(r, w)

	if backref != "" {
		http.Redirect(w, r, backref, http.StatusFound)
	} else {
		http.Redirect(w, r, "/home", http.StatusFound)
	}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r)
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// buildRTMPPushURL constructs the RTMP push URL for OBS.
func buildRTMPPushURL(hostname string, rtmpPort int, streamKey string) string {
	if rtmpPort == 1935 {
		return fmt.Sprintf("rtmp://%s/live/%s", hostname, url.PathEscape(streamKey))
	}
	return fmt.Sprintf("rtmp://%s:%d/live/%s", hostname, rtmpPort, url.PathEscape(streamKey))
}
