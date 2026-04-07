package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/titagaki/pcgw-0yp/internal/model"
)

const userKey contextKey = "user"

// publicPrefixes are paths that don't require authentication.
var publicPrefixes = []string{
	"/login",
	"/auth/",
	"/api/1/",
	"/public/",
	"/stats",
	"/profile/",
	"/programs/",
}

func Auth(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Top page is public
			if path == "/" {
				next.ServeHTTP(w, r)
				return
			}

			// Check public prefixes
			for _, prefix := range publicPrefixes {
				if strings.HasPrefix(path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

			session := GetSession(r)
			uid, ok := session.Values["uid"].(int64)
			if !ok {
				http.Redirect(w, r, "/login?backref="+r.URL.RequestURI(), http.StatusFound)
				return
			}

			user, err := model.GetUser(database, uid)
			if err != nil {
				// Invalid session
				session.Values = make(map[interface{}]interface{})
				session.Save(r, w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			if user.Suspended {
				http.Error(w, "アカウントが凍結されています", http.StatusForbidden)
				return
			}

			model.UpdateUserLoggedOn(database, user.ID)

			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := CurrentUser(r)
		if user == nil || !user.Admin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CurrentUser(r *http.Request) *model.User {
	user, _ := r.Context().Value(userKey).(*model.User)
	return user
}

func IsLoggedIn(r *http.Request) bool {
	session := GetSession(r)
	_, ok := session.Values["uid"].(int64)
	return ok
}
