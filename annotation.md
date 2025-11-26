### What does the application do?

The `swag` application automatically generates documentation in **OpenAPI 3.0** format (formerly known as Swagger) from comments (annotations) in your Go code.

**Main functions:**

1.  **Code Parsing:** It goes through your `.go` files and looks for specific annotations that start with the `@` character.
2.  **Documentation Generation:** Based on the found annotations, it generates `swagger.json` and `swagger.yaml` files that describe your API.
3.  **Integration with Swagger UI:** It also generates `docs.go`, which allows for easy integration of Swagger UI into your Go web application and displays interactive API documentation directly in the browser.
4.  **Framework Support:** It supports popular Go web frameworks like Gin, Echo, Gorilla Mux, and others.

In short, `swag` saves you the work of manually writing and maintaining OpenAPI specifications and ensures that your documentation is always in sync with your code.

### List of available annotations

Here is a list of annotations you can use, divided into two main categories:

#### 1. General API Info

These annotations are usually placed in the `main.go` file (or another main file you specify) and define the global properties of your API.

| Annotation | Description | Example |
| :--- | :--- | :--- |
| **@title** | **Required.** The title of your application. | `// @title My API` |
| **@version** | **Required.** The version of your API. | `// @version 1.0` |
| **@description** | A short description of your application. Can be multi-line. | `// @description This is a sample server.` |
| **@description.markdown** | Loads the description from a markdown file. | `// @description.markdown` |
| **@termsOfService** | A link to the terms of service. | `// @termsOfService http://example.com/terms` |
| **@contact.name** | The name of the contact person/team. | `// @contact.name API Support` |
| **@contact.url** | The URL to the contact page. | `// @contact.url http://www.example.com/support` |
| **@contact.email** | The email address of the contact. | `// @contact.email support@example.com` |
| **@license.name** | **Required.** The name of the license. | `// @license.name Apache 2.0` |
| **@license.url** | A link to the full text of the license. | `// @license.url http://www.apache.org/licenses/LICENSE-2.0.html` |
| **@host** | The host (domain or IP) where the API is available. | `// @host localhost:8080` |
| **@hoststate** | Conditional host setting based on state. Requires 3 arguments: annotation, state value, and host. | `// @hoststate production prod api.example.com` |
| **@BasePath** | The base URL path under which the API is available. | `// @BasePath /api/v1` |
| **@schemes** | A list of transfer protocols (space-separated). | `// @schemes http https` |
| **@accept** | The default MIME types that the API accepts (e.g., `json`, `xml`). | `// @accept json` |
| **@produce** | The default MIME types that the API produces. | `// @produce json` |
| **@externalDocs.description** | A description of the external documentation. | `// @externalDocs.description More information` |
| **@externalDocs.url** | A link to the external documentation. | `// @externalDocs.url https://example.com` |
| **@tag.name** | Defines a tag name for grouping API operations. | `// @tag.name Users` |
| **@tag.description** | A description of the tag. | `// @tag.description Operations related to user management` |
| **@tag.description.markdown** | Loads the tag description from a markdown file named `{tagname}.md`. | `// @tag.description.markdown` |
| **@tag.docs.url** | A URL to external documentation for the tag. | `// @tag.docs.url https://example.com/docs/users` |
| **@tag.docs.description** | A description of the external documentation for the tag. | `// @tag.docs.description External user management docs` |
| **@tag.x-*** | Custom extension fields for tags (e.g., `@tag.x-displayName`). The value is added to tag extensions. | `// @tag.x-displayName User Management` |
| **@query.collection.format** | The default collection (array) param format in query. Possible values: `csv`, `multi`, `pipes`, `tsv`, `ssv`. Default is `csv`. | `// @query.collection.format multi` |
| **@securityDefinitions.basic** | **[Swagger 2.0]** Defines Basic HTTP authentication. | `// @securityDefinitions.basic BasicAuth` |
| **@securityDefinitions.apikey** | **[Swagger 2.0]** Defines authentication using an API key. Requires `in` and `name` parameters. | `// @securitydefinitions.apikey ApiKeyAuth`<br>`// @in header`<br>`// @name X-API-Key` |
| **@securityDefinitions.oauth2.application** | **[Swagger 2.0]** Defines OAuth2 application flow. Requires `tokenUrl`. | `// @securitydefinitions.oauth2.application OAuth2App`<br>`// @tokenUrl https://example.com/oauth/token` |
| **@securityDefinitions.oauth2.implicit** | **[Swagger 2.0]** Defines OAuth2 implicit flow. Requires `authorizationUrl`. | `// @securitydefinitions.oauth2.implicit OAuth2Implicit`<br>`// @authorizationUrl https://example.com/oauth/authorize` |
| **@securityDefinitions.oauth2.password** | **[Swagger 2.0]** Defines OAuth2 password flow. Requires `tokenUrl`. | `// @securitydefinitions.oauth2.password OAuth2Password`<br>`// @tokenUrl https://example.com/oauth/token` |
| **@securityDefinitions.oauth2.accessCode** | **[Swagger 2.0]** Defines OAuth2 access code flow. Requires `tokenUrl` and `authorizationUrl`. | `// @securitydefinitions.oauth2.accessCode OAuth2AccessCode`<br>`// @tokenUrl https://example.com/oauth/token`<br>`// @authorizationUrl https://example.com/oauth/authorize` |
| **@securitySchemes.http** | **[OpenAPI 3.0]** Defines HTTP authentication (Basic or Bearer). Requires `scheme` parameter. | `// @securitySchemes.http BasicAuth`<br>`// @scheme basic` |
| **@securitySchemes.apikey** | **[OpenAPI 3.0]** Defines API key authentication. Requires `in` and `name` parameters. | `// @securitySchemes.apikey ApiKeyAuth`<br>`// @in header`<br>`// @name X-API-Key` |
| **@securitySchemes.oauth2.authorizationCode** | **[OpenAPI 3.0]** Defines OAuth2 authorization code flow. Requires `tokenUrl` and `authorizationUrl`. | `// @securitySchemes.oauth2.authorizationCode OAuth2AuthCode`<br>`// @tokenUrl https://example.com/oauth/token`<br>`// @authorizationUrl https://example.com/oauth/authorize` |
| **@securitySchemes.oauth2.implicit** | **[OpenAPI 3.0]** Defines OAuth2 implicit flow. Requires `authorizationUrl`. | `// @securitySchemes.oauth2.implicit OAuth2Implicit`<br>`// @authorizationUrl https://example.com/oauth/authorize` |
| **@securitySchemes.oauth2.password** | **[OpenAPI 3.0]** Defines OAuth2 password flow. Requires `tokenUrl`. | `// @securitySchemes.oauth2.password OAuth2Password`<br>`// @tokenUrl https://example.com/oauth/token` |
| **@securitySchemes.oauth2.clientCredentials** | **[OpenAPI 3.0]** Defines OAuth2 client credentials flow. Requires `tokenUrl`. | `// @securitySchemes.oauth2.clientCredentials OAuth2ClientCreds`<br>`// @tokenUrl https://example.com/oauth/token` |
| **@securitySchemes.openidconnect** | **[OpenAPI 3.0]** Defines OpenID Connect authentication. Requires `openidConnectUrl`. | `// @securitySchemes.openidconnect OpenIDConnect`<br>`// @openidConnectUrl https://example.com/.well-known/openid-configuration` |
| **@scope.{name}** | Defines a scope for OAuth2 security definitions. Used after security definition. | `// @scope.write Grants write access`<br>`// @scope.read Grants read access` |
| **@x-{name}** | Allows adding custom (extension) fields to the specification. Value must be valid JSON. | `// @x-custom-info {"key": "value"}` |

#### 2. Annotations for individual operations (API Operation)

These annotations are written directly above the functions (handlers) of your endpoints and describe specific operations.

| Annotation | Description | Example |
| :--- | :--- | :--- |
| **@summary** | A short summary of what the operation does. | `// @summary Shows user details` |
| **@description** | A detailed description of the operation. Can be multi-line. Supports `@file` directive to load from file. | `// @description Gets all data about a specific user.`<br>`// @description @file ./docs/user-details.md` |
| **@description.markdown** | Loads the operation description from a markdown file. | `// @description.markdown user-details.md` |
| **@state** | Defines the state for the operation (useful for conditional generation). | `// @state development` |
| **@id** | A unique identifier for the operation. | `// @id getUserById` |
| **@tags** | A list of tags (categories) for grouping operations (comma-separated). | `// @tags Users,Admin` |
| **@accept** | Specific MIME types that this endpoint accepts. | `// @accept json` |
| **@produce** | Specific MIME types that this endpoint returns. | `// @produce xml` |
| **@param** | Describes an endpoint parameter. Format: `name param_type data_type required "description" [attributes]` | `// @param id path int true "User ID"` |
| **@success** | Describes a successful response. Format: `http_code {param_type} data_type "description"` | `// @success 200 {object} model.User "Successfully returned user"` |
| **@failure** | Describes an error response. Format: `http_code {param_type} data_type "description"` | `// @failure 404 {object} httputil.HTTPError "User not found"` |
| **@response** | A generic annotation for a response (can be used instead of `@success` and `@failure`). | `// @response 200 {object} model.User "OK"` |
| **@header** | Defines a header in the response. Format: `http_code {header_type} data_type "description"` | `// @header 200 {string} Token "Authorization token"` |
| **@router** | **Required.** Defines the URL path and HTTP method. Format: `path [method]` | `// @router /users/{id} [get]` |
| **@deprecatedrouter** | **Deprecated.** Same as `@router` but marks the route as deprecated. Format: `path [method]` | `// @deprecatedrouter /users/{id} [get]` |
| **@deprecated** | Marks the endpoint as deprecated. | `// @deprecated` |
| **@private** | Marks the endpoint as private. It will not be included in the public documentation. | `// @private` |
| **@privateFile** | Loads content from a file and appends it to the description only in the private documentation. | `// @privateFile internal_notes.md` |
| **@security** | Applies a security scheme to the given operation. | `// @security BasicAuth` |
| **@x-codeSample** | Adds a code sample. Use `file` to load from an external file named `{operation-summary}.md`. Can also accept JSON directly. | `// @x-codeSample file` |
| **@x-{name}** | Allows adding custom (extension) fields to the operation. Value must be valid JSON. | `// @x-custom-field {"key": "value"}` |