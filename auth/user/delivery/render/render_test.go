package render

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "../templates/"
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Errorf("Error when creating template cache, %v", err)
		return
	}

	if err = RenderTemplate(&fasthttp.RequestCtx{}, fasthttp.StatusOK, tc["login.page.html"], nil); err != nil {
		t.Errorf("Error when rendering template, %v", err)
		return
	}

	if err = RenderTemplate(&fasthttp.RequestCtx{}, fasthttp.StatusOK, tc["non-existent.page.html"], nil); err == nil {
		t.Error("Rendered template that does not exist")
		return
	}
}
