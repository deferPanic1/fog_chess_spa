package httphandler

import (
	"net/http"
	"time"
)

func setAccessTokenCookie(w http.ResponseWriter, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAccessTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteLaxMode,
	})
}
