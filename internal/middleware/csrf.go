package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)

		// Generate token if not present
		if _, ok := session.Values["csrf_token"].(string); !ok {
			b := make([]byte, 32)
			rand.Read(b)
			session.Values["csrf_token"] = hex.EncodeToString(b)
			session.Save(r, w)
		}

		// Skip validation for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		token := r.FormValue("csrf_token")
		expected, _ := session.Values["csrf_token"].(string)
		if token == "" || token != expected {
			http.Error(w, "Forbidden - invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CSRFToken(r *http.Request) string {
	session := GetSession(r)
	token, _ := session.Values["csrf_token"].(string)
	return token
}
