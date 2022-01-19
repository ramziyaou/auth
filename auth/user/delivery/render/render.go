package render

import (
	"auth/myerrors"
	"fmt"
	"log"
	"path/filepath"
	"text/template"

	"github.com/valyala/fasthttp"
)

var pathToTemplates = "./templates/"

var functions = template.FuncMap{
	"inc": func(i int) int {
		return i + 1
	},
}

// RenderTemplate renders a template
func RenderTemplate(ctx *fasthttp.RequestCtx, status int, t *template.Template, data interface{}) error {
	if t == nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return myerrors.ErrNilTemplate
	}
	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType("text/html")
	if err := t.Execute(ctx, data); err != nil {
		return err
	}
	return nil
}

// CreateTemplateCache creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {
	log.Println("INFO|CreateTemplateCache hit")
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s*.page.html", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s*.layout.html", pathToTemplates))
		if err != nil {
			return myCache, err
		}
		if matches == nil {
			log.Println("couldnt find files")
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s*.layout.html", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}
		myCache[name] = ts
	}
	log.Println("INFO|Template cache:", myCache)
	return myCache, nil
}
