package secpkg

import (
	"context"

	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
)

// mockResolver is a mock resolver useful for testing.
type mockResolver struct {
	Files map[string]string
	Heads map[string]ssot.SignedHead
	URLs  map[string][]string
}

func (r *mockResolver) Download(filepath string, url string) error {
	return file.Copy(r.Files[url], filepath)
}

func (r *mockResolver) LookupHead(ctx context.Context, dns string) (ssot.SignedHead, error) {
	sh, ok := r.Heads[dns]
	if ok {
		return sh, nil
	}
	return nil, ssot.ErrTXTNoValidHead
}

func (r *mockResolver) LookupURLs(ctx context.Context, dns string) ([]string, error) {
	urls, ok := r.URLs[dns]
	if ok {
		return urls, nil
	}
	return nil, ssot.ErrTXTNoValidURL
}
