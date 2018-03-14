package prefixhandler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acoshift/prefixhandler"
)

type (
	inj     struct{}
	ctxKey1 struct{}
	ctxKey2 struct{}
)

type injItem struct {
	Key1 string
	Key2 string
	Path string
}

func _handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	*ctx.Value(inj{}).(*injItem) = injItem{
		Key1: prefixhandler.Get(ctx, ctxKey1{}),
		Key2: prefixhandler.Get(ctx, ctxKey2{}),
		Path: r.URL.Path,
	}
}

var (
	handler = http.HandlerFunc(_handler)
	ctxBg   = context.Background()
)

func test(t *testing.T, path, prefix1, prefix2 string, ks int, k1, k2 string, afterPath string) {
	var p injItem
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r = r.WithContext(context.WithValue(ctxBg, inj{}, &p))

	var h http.Handler = handler

	if ks == 2 {
		h = prefixhandler.New(prefix1, ctxKey1{}, prefixhandler.New(prefix2, ctxKey2{}, h))
	} else if ks == 1 {
		h = prefixhandler.New(prefix1, ctxKey1{}, h)
	}

	h.ServeHTTP(w, r)
	if p.Key1 != k1 {
		t.Errorf("invalid key 1 for path %s; expected %s; got %s", path, k1, p.Key1)
	}
	if p.Key2 != k2 {
		t.Errorf("invalid key 2 for path %s; expected %s; got %s", path, k2, p.Key2)
	}
	if p.Path != afterPath {
		t.Errorf("invalid after path for path %s; expected %s; got %s", path, afterPath, p.Path)
	}
}

func TestPrefixHandler(t *testing.T) {
	testCases := []struct {
		path             string
		prefix1, prefix2 string
		ks               int
		k1, k2           string
		afterPath        string
	}{
		{"/", "", "", 0, "", "", "/"},

		{"/", "", "", 1, "", "", "/"},
		{"/", "/", "", 1, "", "", "/"},
		{"/p", "/", "", 1, "p", "", "/"},
		{"/p", "/p", "", 1, "", "", "/"},
		{"/p/1", "/p", "", 1, "1", "", "/"},
		{"/p/1/e", "/p", "", 1, "1", "", "/e"},
		{"/p/1/e/a", "/p", "", 1, "1", "", "/e/a"},
		{"/p/123a/e/a", "/p", "", 1, "123a", "", "/e/a"},
		{"/p/k/1/e/a", "/p/k", "", 1, "1", "", "/e/a"},

		{"/", "", "", 2, "", "", "/"},
		{"/", "/", "", 2, "", "", "/"},
		{"/p", "/", "", 2, "p", "", "/"},
		{"/p", "/p", "", 2, "", "", "/"},
		{"/p/1", "/p", "", 2, "1", "", "/"},
		{"/p/1/e", "/p", "", 2, "1", "e", "/"},
		{"/p/1/e/a", "/p", "/e", 2, "1", "a", "/"},
		{"/p/123a/e/a", "/p", "/e", 2, "123a", "a", "/"},
		{"/p/123a/e/a/l", "/p", "/e", 2, "123a", "a", "/l"},
		{"/p/k/1/e/a/2/l", "/p/k", "/e/a", 2, "1", "2", "/l"},
	}
	for _, c := range testCases {
		test(t, c.path, c.prefix1, c.prefix2, c.ks, c.k1, c.k2, c.afterPath)
	}
}
