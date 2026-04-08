package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/titagaki/pcgw-0yp/internal/handler"
	"github.com/titagaki/pcgw-0yp/internal/middleware"
)

func NewRouter(h *handler.Handler, sessionMiddleware func(http.Handler) http.Handler, authMiddleware func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(sessionMiddleware)
	r.Use(middleware.CSRF)
	r.Use(authMiddleware)

	// Static files
	fileServer := http.FileServer(http.Dir("public"))
	r.Handle("/public/*", http.StripPrefix("/public/", fileServer))

	// Public pages
	r.Get("/", h.TopPage)
	r.Get("/doc/desc", h.DescPage)
	r.Get("/login", h.LoginPage)
	r.Get("/auth/twitter", h.TwitterLogin)
	r.Get("/auth/twitter/callback", h.TwitterCallback)
	r.Get("/logout", h.Logout)

	// Stats (public)
	r.Get("/stats", h.Stats)

	// Public API
	r.Get("/api/1/channelStatus", h.APIChannelStatus)

	// Profile (public)
	r.Get("/profile", h.ProfileList)
	r.Get("/profile/{id}", h.ProfileShow)

	// Programs (public)
	r.Get("/programs", h.ProgramIndex)
	r.Get("/programs/recent", h.ProgramRecent)
	r.Get("/programs/by-date/{year}/{month}", h.ProgramByMonth)
	r.Get("/programs/{id}", h.ProgramShow)

	// Authenticated pages
	r.Get("/home", h.HomePage)

	// Profile edit
	r.Get("/profile/edit", h.ProfileEdit)
	r.Post("/profile/edit", h.ProfileUpdate)

	// Broadcast
	r.Get("/create", h.CreatePage)
	r.Post("/broadcast", h.Broadcast)

	// Channels
	r.Get("/channels", h.ChannelList)
	r.Get("/channels/{id}", h.ChannelShow)
	r.Get("/channels/{id}/edit", h.ChannelEdit)
	r.Post("/channels/{id}", h.ChannelUpdate)
	r.Post("/channels/{id}/stop", h.ChannelStop)
	r.Get("/channels/{id}/relay_tree", h.ChannelRelayTree)
	r.Get("/channels/{id}/connections", h.ChannelConnections)
	r.Post("/channels/{id}/connections/{connID}/disconnect", h.ChannelDisconnect)
	r.Get("/channels/{id}/status.json", h.ChannelStatusJSON)

	// Sources
	r.Get("/sources", h.SourceIndex)
	r.Post("/sources/add", h.SourceAdd)
	r.Post("/sources/del", h.SourceDelete)
	r.Post("/sources/regen", h.SourceRegen)

	// Programs (authenticated)
	r.Post("/programs/{id}/delete", h.ProgramDelete)

	// Admin routes
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.RequireAdmin)
		r.Get("/", h.AdminIndex)
	})

	r.Route("/users", func(r chi.Router) {
		r.Use(middleware.RequireAdmin)
		r.Get("/", h.UserList)
		r.Get("/{id}", h.UserShow)
		r.Get("/{id}/edit", h.UserEdit)
		r.Post("/{id}", h.UserUpdate)
		r.Post("/{id}/delete", h.UserDelete)
	})

	r.Route("/servents", func(r chi.Router) {
		r.Use(middleware.RequireAdmin)
		r.Get("/", h.ServentIndex)
		r.Post("/", h.ServentCreate)
		r.Get("/{id}", h.ServentShow)
		r.Post("/{id}", h.ServentUpdate)
		r.Post("/{id}/delete", h.ServentDelete)
		r.Post("/{id}/refresh", h.ServentRefresh)
	})

	r.Route("/notices", func(r chi.Router) {
		r.Get("/", h.NoticeIndex)
		r.Get("/{id}", h.NoticeShow)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Get("/new", h.NoticeNew)
			r.Post("/", h.NoticeCreate)
			r.Get("/{id}/edit", h.NoticeEdit)
			r.Post("/{id}/update", h.NoticeUpdate)
			r.Post("/{id}/delete", h.NoticeDelete)
		})
	})

	return r
}
