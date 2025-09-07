package main

// @title           Test API pro @file direktivu
// @version         1.0
// @description     @file ./docs/api.md
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @BasePath  /api/v1

func main() {}

// TestEndpoint tests the @file directive for endpoint descriptions
// @Summary      Test endpoint
// @Description  @file ./docs/endpoint.md
// @Tags         test
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /test [get]
func TestEndpoint() {}

// RegularEndpoint tests normal description without @file
// @Summary      Regular endpoint
// @Description  This is a regular description without file directive
// @Tags         test
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /regular [get]
func RegularEndpoint() {}
