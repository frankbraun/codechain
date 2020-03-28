package secpkg

import (
	"context"

	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
)

// Resolver is used for DNS queries and to download file.
type Resolver interface {
	Download(filepath string, url string) error
	LookupHead(ctx context.Context, dns string) (ssot.SignedHead, error)
	LookupURLs(ctx context.Context, dns string) ([]string, error)
}

type stdResolver struct{}

func (r stdResolver) Download(filepath string, url string) error {
	return file.Download(filepath, url)
}

func (r stdResolver) LookupHead(ctx context.Context, dns string) (ssot.SignedHead, error) {
	return ssot.LookupHead(ctx, dns)
}

func (r stdResolver) LookupURLs(ctx context.Context, dns string) ([]string, error) {
	return ssot.LookupURLs(ctx, dns)
}

// NewResolver returns a new standard resolver.
func NewResolver() Resolver {
	return stdResolver{}
}
