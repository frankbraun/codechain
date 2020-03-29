package secpkg

import (
	"context"
	"fmt"

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
	fmt.Printf("mockResolver.Download(%s, %s)\n", filepath, url)
	fn, ok := r.Files[url]
	if ok {
		fmt.Printf("mockResolver.Download: file.Copy(%s, %s)\n", fn, filepath)
		return file.Copy(fn, filepath)
	}
	return fmt.Errorf("mockResolver: file %s not found", url)
}

func (r *mockResolver) LookupHead(ctx context.Context, dns string) (ssot.SignedHead, error) {
	fmt.Printf("mockResolver.LookupHead(ctx, %s)\n", dns)
	sh, ok := r.Heads[dns]
	if ok {
		return sh, nil
	}
	return nil, ssot.ErrTXTNoValidHead
}

func (r *mockResolver) LookupURLs(ctx context.Context, dns string) ([]string, error) {
	fmt.Printf("mockResolver.LookupURLs(ctx, %s)\n", dns)
	urls, ok := r.URLs[dns]
	if ok {
		return urls, nil
	}
	return nil, ssot.ErrTXTNoValidURL
}
