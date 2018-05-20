package micro

import (
	"bytes"
	"html/template"
	"net/http"
)

// RedocOpts is the Redoc configures type
type RedocOpts struct {
	// SpecURL the url to find the spec for
	SpecURL string
	// RedocURL for the js that generates the redoc site, defaults to: https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js
	RedocURL string
	// Title for the documentation site, default to: API documentation
	Title string
}

// RedocOptions configures the Redoc
var RedocOptions RedocOpts

func (r *RedocOpts) ensureDefaults() {
	if r.SpecURL == "" {
		r.SpecURL = "/swagger.json"
	}
	if r.RedocURL == "" {
		r.RedocURL = "https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js"
	}
	if r.Title == "" {
		r.Title = "API documentation"
	}
}

// redoc is the HandlerFunc for Redoc
func redoc(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {

	RedocOptions.ensureDefaults()

	tmpl := template.Must(template.New("redoc").Parse(redocTemplate))

	buf := bytes.NewBuffer(nil)
	_ = tmpl.Execute(buf, RedocOptions)
	b := buf.Bytes()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

const (
	redocTemplate = `<!DOCTYPE html>
<html>
  <head>
    <title>{{ .Title }}</title>
    <!-- needed for adaptive design -->
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <!--
    ReDoc doesn't change outer page styles
    -->
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc spec-url='{{ .SpecURL }}'></redoc>
    <script src="{{ .RedocURL }}"> </script>
  </body>
</html>
`
)
