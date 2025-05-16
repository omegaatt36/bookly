package web

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt"
)

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	loginData := map[string]string{
		"email":    username,
		"password": password,
	}

	var loginResp struct {
		Token string `json:"token"`
	}

	if err := s.sendRequest(r, "POST", "/public/auth/login", loginData, &loginResp); err != nil {
		http.Error(w, "Login failed", http.StatusUnauthorized)
		return
	}

	// Set the token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    loginResp.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) logout(w http.ResponseWriter, _ *http.Request) {
	s.clearTokenAndRedirect(w)
}

func (s *Server) clearTokenAndRedirect(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getUserIDFromToken(tokenString string) (int32, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userIDFloat, ok := claims["user_id"].(float64); ok {
			return int32(userIDFloat), nil
		}
	}

	return 0, errors.New("user_id not found in token claims")
}
