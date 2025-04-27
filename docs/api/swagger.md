# Routes and Methods

GoToSocial uses [go-swagger](https://github.com/go-swagger/go-swagger) to generate a V2 [OpenAPI specification](https://swagger.io/specification/v2/) document from code annotations.

The resulting API documentation is rendered below. Please note that the doc is intended for reference only. You will not be able to use the built-in Authorize functionality in the below widget to actually connect to an instance or make API calls. Instead, you should use something like curl, Postman, or similar.

Most of the GoToSocial API endpoints require a user-level OAuth token. For a guide on how to authenticate with the API using an OAuth token, see the [authentication doc](./authentication.md).

!!! tip
    If you'd like to do more with the spec, you can also view the [swagger.yaml](./swagger.yaml) directly, and then paste it into something like the [Swagger Editor](https://editor.swagger.io/). That way you can try autogenerating GoToSocial API clients in different languages (not supported, but fun), or convert the doc to JSON or OpenAPI v3 specification, etc. See [here](https://swagger.io/tools/open-source/getting-started/) for more.

!!! info "Gotcha: uploading files"
    When using an API endpoint that involves submitting a form to upload files (eg., the media attachments endpoints, or the emoji upload endpoint, etc), please note that `filename` is a required on the form field, due to the dependency that GoToSocial uses to parse forms, and some quirks in Go.
    
    See the following issues for more context:
    
    - [#1958](https://codeberg.org/superseriousbusiness/gotosocial/issues/1958)
    - [#1944](https://codeberg.org/superseriousbusiness/gotosocial/issues/1944)
    - [#2641](https://codeberg.org/superseriousbusiness/gotosocial/issues/2641)

<swagger-ui src="swagger.yaml"/>
