// Package docs menyediakan halaman login untuk akses dokumentasi API.
package docs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	cookieName    = "docs_session"
	cookieMaxAge  = 3600 // 1 hour
	sessionSecret = "docs-session-secret-change-in-production"
)

// LoginHandler menangani GET (form) dan POST (submit) login docs.
type LoginHandler struct {
	DocsUser string
	DocsPass string
}

// NewLoginHandler membuat handler login docs.
func NewLoginHandler(user, pass string) *LoginHandler {
	return &LoginHandler{DocsUser: user, DocsPass: pass}
}

// ServeLogin menampilkan form login (GET) dan memproses submit (POST).
func (h *LoginHandler) ServeLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.handleLogin(w, r)
		return
	}
	// GET: tampilkan form
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(loginHTML))
}

func (h *LoginHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username != h.DocsUser || password != h.DocsPass {
		// Redirect back to login with error
		http.Redirect(w, r, "/docs/login?error=1", http.StatusSeeOther)
		return
	}

	// Create session token (HMAC-signed)
	token := createSessionToken(username)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/docs", http.StatusSeeOther)
}

// ServeLogout menghapus cookie session.
func (h *LoginHandler) ServeLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/docs/login", http.StatusSeeOther)
}

// DocsAuthMiddleware memeriksa cookie session sebelum serve /docs & /openapi.json.
func DocsAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for non-docs paths
		if r.URL.Path != "/docs" && r.URL.Path != "/openapi.json" && r.URL.Path != "/schemas" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for login page
		if r.URL.Path == "/docs/login" || r.URL.Path == "/docs/logout" {
			next.ServeHTTP(w, r)
			return
		}

		// Check cookie
		cookie, err := r.Cookie(cookieName)
		if err != nil || !validateSessionToken(cookie.Value) {
			http.Redirect(w, r, "/docs/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// createSessionToken membuat HMAC-signed session token.
func createSessionToken(username string) string {
	expiry := time.Now().Add(time.Duration(cookieMaxAge) * time.Second).Unix()
	data := fmt.Sprintf("%s:%d", username, expiry)
	mac := hmac.New(sha256.New, []byte(sessionSecret))
	mac.Write([]byte(data))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	return base64.URLEncoding.EncodeToString([]byte(data)) + "." + sig
}

// validateSessionToken memvalidasi HMAC-signed session token.
func validateSessionToken(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	data, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	sig, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	// Verify HMAC
	mac := hmac.New(sha256.New, []byte(sessionSecret))
	mac.Write(data)
	expectedSig := mac.Sum(nil)
	if !hmac.Equal(sig, expectedSig) {
		return false
	}

	// Check expiry
	dataStr := string(data)
	lastColon := strings.LastIndex(dataStr, ":")
	if lastColon == -1 {
		return false
	}
	expiryStr := dataStr[lastColon+1:]
	var expiry int64
	_, err = fmt.Sscanf(expiryStr, "%d", &expiry)
	if err != nil {
		return false
	}
	if time.Now().Unix() > expiry {
		return false
	}

	return true
}

// loginHTML is the login page HTML (embedded).
var loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta content="width=device-width, initial-scale=1.0" name="viewport"/>
<title>KicauGo Docs Login</title>
<script src="https://cdn.tailwindcss.com"></script>
<style>
  body { font-family: 'Inter', sans-serif; }
</style>
</head>
<body class="min-h-screen flex items-center justify-center bg-gray-100">
  <div class="w-full max-w-md">
    <div class="bg-white rounded-2xl shadow-xl p-8">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-bold text-gray-900">KicauGo Docs</h1>
        <p class="text-gray-500 mt-2">API Documentation Access</p>
      </div>
      <div id="error" class="hidden mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
        Invalid username or password
      </div>
      <form method="POST" action="/docs/login" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Username</label>
          <input type="text" name="username" required
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="Enter username">
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input type="password" name="password" required
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="Enter password">
        </div>
        <button type="submit"
          class="w-full py-3 px-4 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-lg transition-colors">
          Login
        </button>
      </form>
    </div>
  </div>
  <script>
    // Show error if redirected with ?error=1
    if (new URLSearchParams(window.location.search).get('error')) {
      document.getElementById('error').classList.remove('hidden');
    }
  </script>
</body>
</html>`
