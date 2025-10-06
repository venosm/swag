package main

import (
	"net/http"

	"github.com/venosm/swag/testdata/openapi3_security/api"
)

// @title Swagger Example API with OpenAPI 3.0 Security
// @version 1.0
// @description This is a sample server demonstrating OpenAPI 3.0 security schemes.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /v1

// @securitySchemes.http BasicAuth
// @scheme basic
// @description HTTP Basic Authentication
//
// @securitySchemes.http BearerAuth
// @scheme bearer
// @bearerFormat JWT
// @description JWT Bearer token authentication
//
// @securitySchemes.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key authentication
//
// @securitySchemes.oauth2.authorizationCode OAuth2
// @tokenUrl https://example.com/oauth/token
// @authorizationUrl https://example.com/oauth/authorize
// @scope.read Grants read access
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information
//
// @securitySchemes.openidconnect OpenIDConnect
// @openidConnectUrl https://example.com/.well-known/openid-configuration
// @description OpenID Connect Discovery

func main() {
	http.HandleFunc("/v1/public", api.GetPublic)
	http.HandleFunc("/v1/basic", api.GetBasic)
	http.HandleFunc("/v1/bearer", api.GetBearer)
	http.HandleFunc("/v1/apikey", api.GetAPIKey)
	http.HandleFunc("/v1/oauth", api.GetOAuth)
	http.HandleFunc("/v1/openid", api.GetOpenID)
	http.ListenAndServe(":8080", nil)
}
