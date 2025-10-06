package api

import (
	"net/http"
)

// GetPublic godoc
// @Summary Public endpoint
// @Description This endpoint does not require authentication
// @Tags public
// @Accept json
// @Produce json
// @Success 200 {string} string "ok"
// @Router /public [get]
func GetPublic(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("public"))
}

// GetBasic godoc
// @Summary Basic auth endpoint
// @Description This endpoint requires HTTP Basic authentication
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {string} string "ok"
// @Failure 401 {string} string "unauthorized"
// @Security BasicAuth
// @Router /basic [get]
func GetBasic(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("basic"))
}

// GetBearer godoc
// @Summary Bearer token endpoint
// @Description This endpoint requires JWT Bearer token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {string} string "ok"
// @Failure 401 {string} string "unauthorized"
// @Security BearerAuth
// @Router /bearer [get]
func GetBearer(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("bearer"))
}

// GetAPIKey godoc
// @Summary API key endpoint
// @Description This endpoint requires API key in header
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {string} string "ok"
// @Failure 401 {string} string "unauthorized"
// @Security ApiKeyAuth
// @Router /apikey [get]
func GetAPIKey(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("apikey"))
}

// GetOAuth godoc
// @Summary OAuth2 endpoint
// @Description This endpoint requires OAuth2 authentication
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {string} string "ok"
// @Failure 401 {string} string "unauthorized"
// @Security OAuth2[read, write]
// @Router /oauth [get]
func GetOAuth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("oauth"))
}

// GetOpenID godoc
// @Summary OpenID Connect endpoint
// @Description This endpoint requires OpenID Connect authentication
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {string} string "ok"
// @Failure 401 {string} string "unauthorized"
// @Security OpenIDConnect
// @Router /openid [get]
func GetOpenID(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("openid"))
}
