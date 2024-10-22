package proxymanager;

import (
  "context"
)

type ProxyClient interface {
	StartProxyMapping(ctx context.Context, subdomain string, port int) error
	EndProxyMapping( ctx context.Context, subdomain string ) error
}
