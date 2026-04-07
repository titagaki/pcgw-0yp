package handler

import (
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/titagaki/pcgw-0yp/internal/config"
	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/model"
	"github.com/titagaki/pcgw-0yp/internal/peercast"
	"github.com/titagaki/pcgw-0yp/internal/view"
)

type Handler struct {
	DB        *sql.DB
	Config    *config.Config
	Templates *template.Template
	Log       *slog.Logger
}

func New(db *sql.DB, cfg *config.Config, log *slog.Logger, templateDir string) *Handler {
	tmpl := template.New("").Funcs(view.FuncMap())
	tmpl = template.Must(tmpl.ParseGlob(filepath.Join(templateDir, "*.html")))
	tmpl = template.Must(tmpl.ParseGlob(filepath.Join(templateDir, "partials", "*.html")))

	return &Handler{
		DB:        db,
		Config:    cfg,
		Templates: tmpl,
		Log:       log,
	}
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["User"] = middleware.CurrentUser(r)
	data["LoggedIn"] = middleware.IsLoggedIn(r)
	data["CSRFToken"] = middleware.CSRFToken(r)

	// Check for unread notices
	if user := middleware.CurrentUser(r); user != nil {
		unread, _ := model.HasUnreadNotices(h.DB, user.ID)
		data["HasUnreadNotices"] = unread
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, name, data); err != nil {
		h.Log.Error("template render error", "template", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) flash(w http.ResponseWriter, r *http.Request, msg string) {
	session := middleware.GetSession(r)
	session.AddFlash(msg)
	session.Save(r, w)
}

func (h *Handler) getFlashes(r *http.Request, w http.ResponseWriter) []interface{} {
	session := middleware.GetSession(r)
	flashes := session.Flashes()
	if len(flashes) > 0 {
		session.Save(r, w)
	}
	return flashes
}

func (h *Handler) peercastClient(s *model.Servent) *peercast.Client {
	return peercast.NewClient(s.Hostname, s.Port, s.AuthID, s.Passwd)
}
