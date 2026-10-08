// Package apidocs exposes the API contract and interactive documentation assets.
package apidocs

import _ "embed"

// OpenAPISpec is the API contract served at /openapi.yaml.
//
//go:embed openapi.yaml
var OpenAPISpec []byte

// ScalarHTML renders Scalar against the OpenAPI document served by this API.
const ScalarHTML = `<!doctype html>
<html lang="en">
  <head>
    <title>Bit Learning API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <div id="app"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/openapi.yaml',
        theme: 'default',
        layout: 'modern',
        hideModels: false,
        defaultHttpClient: { targetKey: 'shell', clientKey: 'curl' }
      })
    </script>
  </body>
</html>
`
