package redis

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// ServeDNS implements the plugin.Handler interface.
func (redis Redis) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	zone := plugin.Zones(redis.Zones).Matches(state.Name())
	if zone == "" {
		return plugin.NextOrFailure(redis.Name(), redis.Next, ctx, w, r)
	}

	var (
		records, extra []dns.RR
		truncated      bool
		err            error
	)

	switch state.QType() {
	case dns.TypeA:
		records, truncated, err = redis.A(ctx, zone, state, nil)
	case dns.TypeAAAA:
		records, truncated, err = redis.AAAA(ctx, zone, state, nil)
	case dns.TypeCNAME:
		records, err = redis.CNAME(ctx, zone, state)
	case dns.TypeTXT:
		records, truncated, err = redis.TXT(ctx, zone, state, nil)
	case dns.TypeNS:
		records, truncated, err = redis.NS(ctx, zone, state, nil)
	case dns.TypeMX:
		records, truncated, err = redis.MX(ctx, zone, state, nil)
	case dns.TypeSRV:
		records, truncated, err = redis.SRV(ctx, zone, state, nil)
	case dns.TypePTR:
		records, err = redis.PTR(ctx, zone, state)
	case dns.TypeSOA:
		records, err = redis.SOA(ctx, zone, state)
	case dns.TypeCAA:
		records, err = redis.CAA(ctx, zone, state)

	default:
		return redis.errorANSWER(ctx, state.QName(), dns.RcodeNotImplemented, state, nil)
	}

	if err != nil {
		if err == errKeyNotFound && redis.Fall.Through(state.Name()) {
			return plugin.NextOrFailure(redis.Name(), redis.Next, ctx, w, r)
		}
		// Make err nil when returning here, so we don't log spam for NXDOMAIN.
		return redis.errorANSWER(ctx, state.QName(), dns.RcodeServerFailure, state, nil)
	}

	if len(records) == 0 {
		return redis.errorANSWER(ctx, state.QName(), dns.RcodeSuccess, state, nil)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.Truncated = true, truncated

	m.Answer = append(m.Answer, records...)

	m.Extra = append(m.Extra, extra...)

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (Redis) Name() string { return "redis" }

func (r Redis) errorANSWER(ctx context.Context, zone string, rcode int, state request.Request, err error) (int, error) {

	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative = true
	stateNew := state.NewWithQuestion(state.Name(), dns.TypeSOA)
	m.Ns, _ = r.SOA(ctx, zone, stateNew)
	state.W.WriteMsg(m)
	// Return success as the rcode to signal we have written to the client.
	return dns.RcodeSuccess, err
}
