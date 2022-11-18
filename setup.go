package redis

import (
	"crypto/tls"
	"strconv"
	"strings"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	mwtls "github.com/coredns/coredns/plugin/pkg/tls"
	"github.com/coredns/coredns/plugin/pkg/upstream"
	redisV8 "github.com/go-redis/redis/v8"
)

// go-redis有默认地址
// const defaultAddress = ":6379"

func init() { plugin.Register("redis", setup) }

func setup(c *caddy.Controller) error {
	r, err := redisParse(c)
	if err != nil {
		return plugin.Error("redis", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		r.Next = next
		return r
	})
	return nil
}

func redisParse(c *caddy.Controller) (*Redis, error) {
	redis := Redis{
		KeyPrefix: "",
	}
	var (
		tlsConfig *tls.Config
		err       error
		// addresses      = []string{defaultAddress}
		addresses      []string
		username       string
		password       string
		connectTimeout int
		readTimeout    int
	)

	redis.Upstream = upstream.New()

	for c.Next() {
		redis.Zones = plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)
		for c.NextBlock() {
			switch c.Val() {
			case "fallthrough":
				redis.Fall.SetZonesFromArgs(c.RemainingArgs())
			case "addresses":
				addresses = c.RemainingArgs()
				if len(addresses) == 0 {
					return &Redis{}, c.ArgErr()
				}

			case "username":
				if !c.NextArg() {
					return &Redis{}, c.ArgErr()
				}
				username = c.Val()

			case "password":
				if !c.NextArg() {
					return &Redis{}, c.ArgErr()
				}
				password = c.Val()

			case "key_prefix":
				if !c.NextArg() {
					return &Redis{}, c.ArgErr()
				}
				redis.KeyPrefix = c.Val()
				if strings.HasSuffix(redis.KeyPrefix, ":") {
					redis.KeyPrefix = strings.Trim(redis.KeyPrefix, ":")
				}

			case "tls": // cert key cacertfile
				args := c.RemainingArgs()
				tlsConfig, err = mwtls.NewTLSConfigFromArgs(args...)
				if err != nil {
					return &Redis{}, err
				}

			case "connect_timeout":
				if !c.NextArg() {
					return &Redis{}, c.ArgErr()
				}
				connectTimeout, _ = strconv.Atoi(c.Val())

			case "read_timeout":
				if !c.NextArg() {
					return &Redis{}, c.ArgErr()
				}
				readTimeout, _ = strconv.Atoi(c.Val())

			default:
				if c.Val() != "}" {
					return &Redis{}, c.Errf("unknown property '%s'", c.Val())
				}
			}

		}
	}
	redis.Client = redisV8.NewUniversalClient(&redisV8.UniversalOptions{
		Addrs:       addresses,
		Username:    username,
		Password:    password,
		DialTimeout: time.Second * time.Duration(connectTimeout),
		ReadTimeout: time.Duration(readTimeout) * time.Second,
		TLSConfig:   tlsConfig,
	})

	return &redis, nil

}
