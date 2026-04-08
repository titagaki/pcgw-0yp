package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strings"
	"time"

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
	"/profile",
	"/programs",
	"/doc/",
}

func Auth(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Determine if this is a public path
			isPublic := path == "/"
			if !isPublic {
				for _, prefix := range publicPrefixes {
					if strings.HasPrefix(path, prefix) {
						isPublic = true
						break
					}
				}
			}

			session := GetSession(r)
			uid, ok := session.Values["uid"].(int64)
			if !ok {
				if isPublic {
					next.ServeHTTP(w, r)
					return
				}
				http.Redirect(w, r, "/login?backref="+url.QueryEscape(r.URL.RequestURI()), http.StatusFound)
				return
			}

			user, err := model.GetUser(database, uid)
			if err != nil {
				if isPublic {
					next.ServeHTTP(w, r)
					return
				}
				// Invalid session
				session.Values = make(map[interface{}]interface{})
				session.Save(r, w)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			if user.Suspended {
				if !isPublic {
					http.Error(w, "アカウントが凍結されています", http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			if !user.LoggedOnAt.Valid || time.Since(user.LoggedOnAt.Time) > 5*time.Minute {
				model.UpdateUserLoggedOn(database, user.ID)
			}

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
