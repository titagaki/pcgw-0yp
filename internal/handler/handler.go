package handler

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/titagaki/pcgw-0yp/internal/config"
	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/model"
	"github.com/titagaki/pcgw-0yp/internal/peercast"
	"github.com/titagaki/pcgw-0yp/internal/view"

	"github.com/a-h/templ"
)

type Handler struct {
	DB     *sql.DB
	Config *config.Config
	Log    *slog.Logger
}

func New(db *sql.DB, cfg *config.Config, log *slog.Logger) *Handler {
	return &Handler{
		DB:     db,
		Config: cfg,
		Log:    log,
	}
}

func (h *Handler) pageData(r *http.Request, w http.ResponseWriter) view.PageData {
	pd := view.PageData{
		User:      middleware.CurrentUser(r),
		LoggedIn:  middleware.IsLoggedIn(r),
		CSRFToken: middleware.CSRFToken(r),
	}

	if pd.User != nil {
		pd.HasUnreadNotices, _ = model.HasUnreadNotices(h.DB, pd.User.ID)
	}

	pd.Flashes = h.getFlashStrings(r, w)
	return pd
}

func (h *Handler) renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		h.Log.Error("template render error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) flash(w http.ResponseWriter, r *http.Request, msg string) {
	session := middleware.GetSession(r)
	session.AddFlash(msg)
	session.Save(r, w)
}

func (h *Handler) getFlashStrings(r *http.Request, w http.ResponseWriter) []string {
	session := middleware.GetSession(r)
	flashes := session.Flashes()
	if len(flashes) > 0 {
		session.Save(r, w)
	}
	var result []string
	for _, f := range flashes {
		if s, ok := f.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func (h *Handler) peercastClient(s *model.Servent) *peercast.Client {
	return peercast.NewClient(s.Hostname, s.Port, s.AuthID, s.Passwd)
}
