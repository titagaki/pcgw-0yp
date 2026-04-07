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
	DB          *sql.DB
	Config      *config.Config
	Templates   map[string]*template.Template
	TemplateDir string
	Log         *slog.Logger
}

func New(db *sql.DB, cfg *config.Config, log *slog.Logger, templateDir string) *Handler {
	h := &Handler{
		DB:          db,
		Config:      cfg,
		TemplateDir: templateDir,
		Log:         log,
	}
	h.Templates = h.parseTemplates()
	return h
}

func (h *Handler) parseTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template)

	partials := filepath.Join(h.TemplateDir, "partials", "*.html")
	layout := filepath.Join(h.TemplateDir, "layout.html")

	pages, err := filepath.Glob(filepath.Join(h.TemplateDir, "*.html"))
	if err != nil {
		panic(err)
	}

	for _, page := range pages {
		name := filepath.Base(page)
		if name == "layout.html" {
			continue
		}
		tmpl := template.New("").Funcs(view.FuncMap())
		tmpl = template.Must(tmpl.ParseGlob(partials))
		tmpl = template.Must(tmpl.ParseFiles(layout, page))
		templates[name] = tmpl
	}
	return templates
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
	tmpl, ok := h.Templates[name]
	if !ok {
		h.Log.Error("template not found", "template", name)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
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
