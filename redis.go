package redis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/fall"
	"github.com/coredns/coredns/plugin/pkg/upstream"
	"github.com/coredns/coredns/request"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/miekg/dns"
)

type Redis struct {
	Next   plugin.Handler
	Client redisV8.UniversalClient

	KeyPrefix string
	Zones     []string

	Fall fall.F

	Upstream *upstream.Upstream
}

func (r *Redis) get(ctx context.Context, key, field string) (val string, err error) {
	val, err = r.Client.HGet(ctx, key, field).Result()

	// if err == redisV8.Nil {
	// 	val, err = r.Client.HGet(ctx, AnyKey(key), field).Result()
	// }

	if err == redisV8.Nil {
		err = errKeyNotFound
	}
	return
}

func (r *Redis) cnameGet(ctx context.Context, key string) (rCNAME RecordCNANE, err error) {

	val, err := r.get(ctx, key, dns.Type(dns.TypeCNAME).String())
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(val), &rCNAME)
	if err != nil {
		return nil, err
	}

	return
}

func (r *Redis) Lookup(ctx context.Context, state request.Request, name string) (*dns.Msg, error) {
	return r.Upstream.Lookup(ctx, state, name, state.QType())
}

var errKeyNotFound = errors.New("key not found")

// MinTTL returns the minimal TTL.
func (*Redis) MinTTL(state request.Request) uint32 {
	return 30
}
