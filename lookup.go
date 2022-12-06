package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coredns/coredns/plugin/pkg/dnsutil"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

func (r Redis) A(ctx context.Context, zone string, state request.Request, previousRecords []dns.RR) (records []dns.RR, truncated bool, err error) {
	key := Key(state.Name(), r.KeyPrefix)
doSearch:
	clog.Error(key, " A start: ", time.Now())
	defer clog.Error(key, " A end: ", time.Now())
	val, err := r.get(ctx, key, state.Type())
	switch err {
	case nil:
		var rA RecordA
		err = json.Unmarshal([]byte(val), &rA)
		if err != nil {
			return nil, false, err
		}
		for _, item := range rA {
			records = append(records, item.NewA(state.QName()))
		}

	case errKeyNotFound:
		stateNew := state.NewWithQuestion(state.Name(), dns.TypeCNAME)
		rCNAME, err := r.cname(ctx, zone, stateNew)

		if err != nil {
			if err == errKeyNotFound && !IsAnyKey(key) {
				key = AnyKey(key)
				goto doSearch
			}

			return nil, false, err
		}

		for _, item := range rCNAME {
			if len(previousRecords) > 7 {
				break
			}
			cnameRecode := item.NewCNAME(state.QName())
			if dnsutil.DuplicateCNAME(cnameRecode, previousRecords) {
				continue
			}

			if zone == "." || dns.IsSubDomain(zone, dns.Fqdn(item.Host)) {
				stateNew = state.NewWithQuestion(item.Host, state.QType())
				nextRecords, tc, err := r.A(ctx, zone, stateNew, append(previousRecords, cnameRecode))
				if tc {
					truncated = true
				}

				if err == nil {
					if len(nextRecords) > 0 {
						records = append(records, cnameRecode)
						records = append(records, nextRecords...)
					}
					continue
				}

				if err != errKeyNotFound && zone != "." {
					continue
				}
			}

			m1, e1 := r.Lookup(ctx, state, cnameRecode.Target)
			if e1 != nil {
				continue
			}
			if m1.Truncated {
				truncated = true
			}

			records = append(records, cnameRecode)
			records = append(records, m1.Answer...)
		}

	default:
		if err != nil {
			return nil, false, err
		}
	}

	return records, truncated, nil
}

func (r Redis) AAAA(ctx context.Context, zone string, state request.Request, previousRecords []dns.RR) (records []dns.RR, truncated bool, err error) {
	key := Key(state.Name(), r.KeyPrefix)
doSearch:
	val, err := r.get(ctx, key, state.Type())
	switch err {
	case nil:
		var rAAAA RecordAAAA
		err = json.Unmarshal([]byte(val), &rAAAA)
		if err != nil {
			return nil, false, err
		}
		for _, item := range rAAAA {
			records = append(records, item.NewAAAA(state.QName()))
		}

	case errKeyNotFound:
		stateNew := state.NewWithQuestion(state.Name(), dns.TypeCNAME)
		rCNAME, err := r.cname(ctx, zone, stateNew)

		if err != nil {
			if err == errKeyNotFound && !IsAnyKey(key) {
				key = AnyKey(key)
				goto doSearch
			}
			return nil, false, err
		}

		for _, item := range rCNAME {
			if len(previousRecords) > 7 {
				break
			}
			cnameRecode := item.NewCNAME(state.QName())
			if dnsutil.DuplicateCNAME(cnameRecode, previousRecords) {
				continue
			}

			if zone == "." || dns.IsSubDomain(zone, dns.Fqdn(item.Host)) {
				stateNew = state.NewWithQuestion(item.Host, state.QType())
				nextRecords, tc, err := r.AAAA(ctx, zone, stateNew, append(previousRecords, cnameRecode))
				if tc {
					truncated = true
				}

				if err == nil {
					if len(nextRecords) > 0 {
						records = append(records, cnameRecode)
						records = append(records, nextRecords...)
					}
					continue
				}

				if err != errKeyNotFound && zone != "." {
					continue
				}
			}

			m1, e1 := r.Lookup(ctx, state, cnameRecode.Target)
			if e1 != nil {
				continue
			}
			if m1.Truncated {
				truncated = true
			}
			records = append(records, cnameRecode)
			records = append(records, m1.Answer...)
		}

	default:
		if err != nil {
			return nil, false, err
		}
	}
	return records, truncated, nil
}

func (r Redis) CNAME(ctx context.Context, zone string, state request.Request) (records []dns.RR, err error) {
	var rCNAME RecordCNANE

	key := Key(state.Name(), r.KeyPrefix)
doSearch:
	val, err := r.get(ctx, key, state.Type())
	if err != nil {
		if err == errKeyNotFound && !IsAnyKey(key) {
			key = AnyKey(key)
			goto doSearch
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(val), &rCNAME)
	if err != nil {
		return nil, err
	}

	for _, item := range rCNAME {
		records = append(records, item.NewCNAME(state.QName()))
	}
	return
}

func (r Redis) TXT(ctx context.Context, zone string, state request.Request, previousRecords []dns.RR) (records []dns.RR, truncated bool, err error) {
	key := Key(state.Name(), r.KeyPrefix)
doSearch:
	val, err := r.get(ctx, key, state.Type())
	switch err {
	case nil:
		var rTXT RecordTXT
		err = json.Unmarshal([]byte(val), &rTXT)
		if err != nil {
			return nil, false, err
		}
		for _, item := range rTXT {
			records = append(records, item.NewTXT(state.QName()))
		}

	case errKeyNotFound:
		stateNew := state.NewWithQuestion(state.Name(), dns.TypeCNAME)
		rCNAME, err := r.cname(ctx, zone, stateNew)
		if err != nil {
			if err == errKeyNotFound && !IsAnyKey(key) {
				key = AnyKey(key)
				goto doSearch
			}
			return nil, false, err
		}

		for _, item := range rCNAME {

			if len(previousRecords) > 7 {
				break
			}
			cnameRecode := item.NewCNAME(state.QName())
			if dnsutil.DuplicateCNAME(cnameRecode, previousRecords) {
				continue
			}

			if zone == "." || dns.IsSubDomain(zone, dns.Fqdn(item.Host)) {

				stateNew = state.NewWithQuestion(item.Host, state.QType())
				nextRecords, tc, err := r.TXT(ctx, zone, stateNew, append(previousRecords, cnameRecode))
				if tc {
					truncated = true
				}

				if err == nil {
					if len(nextRecords) > 0 {
						records = append(records, cnameRecode)
						records = append(records, nextRecords...)
					}
					continue
				}

				if err != errKeyNotFound && zone != "." {
					continue
				}
			}

			m1, e1 := r.Lookup(ctx, state, cnameRecode.Target)
			if e1 != nil {
				continue
			}
			if m1.Truncated {
				truncated = true
			}
			records = append(records, cnameRecode)
			records = append(records, m1.Answer...)
		}

	default:
		if err != nil {
			return nil, false, err
		}
	}
	return records, truncated, nil
}

func (r Redis) NS(ctx context.Context, zone string, state request.Request, previousRecords []dns.RR) (records []dns.RR, truncated bool, err error) {
	key := Key(state.Name(), r.KeyPrefix)
doSearch:

	val, err := r.get(ctx, key, state.Type())
	switch err {
	case nil:
		var rNS RecordNS
		err = json.Unmarshal([]byte(val), &rNS)
		if err != nil {
			return nil, false, err
		}
		for _, item := range rNS {
			records = append(records, item.NewNS(state.QName()))
		}

	case errKeyNotFound:
		stateNew := state.NewWithQuestion(state.Name(), dns.TypeCNAME)
		rCNAME, err := r.cname(ctx, zone, stateNew)

		if err != nil {
			if err == errKeyNotFound && !IsAnyKey(key) {
				key = AnyKey(key)
				goto doSearch
			}
			return nil, false, err
		}

		for _, item := range rCNAME {
			if len(previousRecords) > 7 {
				break
			}
			cnameRecode := item.NewCNAME(state.QName())
			if dnsutil.DuplicateCNAME(cnameRecode, previousRecords) {
				continue
			}

			if zone == "." || dns.IsSubDomain(zone, dns.Fqdn(item.Host)) {
				stateNew = state.NewWithQuestion(item.Host, state.QType())
				nextRecords, tc, err := r.NS(ctx, zone, stateNew, append(previousRecords, cnameRecode))
				if tc {
					truncated = true
				}

				if err == nil {
					if len(nextRecords) > 0 {
						records = append(records, cnameRecode)
						records = append(records, nextRecords...)
					}
					continue
				}

				if err != errKeyNotFound && zone != "." {
					continue
				}
			}

			m1, e1 := r.Lookup(ctx, state, cnameRecode.Target)
			if e1 != nil {
				continue
			}
			if m1.Truncated {
				truncated = true
			}

			records = append(records, cnameRecode)
			records = append(records, m1.Answer...)
		}

	default:
		if err != nil {
			return nil, false, err
		}
	}
	return records, truncated, nil
}

func (r Redis) PTR(ctx context.Context, zone string, state request.Request) (records []dns.RR, err error) {
	key := Key(state.Name(), r.KeyPrefix)

	val, err := r.get(ctx, key, state.Type())
	if err == nil {
		var rPTR RecordPTR
		err = json.Unmarshal([]byte(val), &rPTR)
		if err == nil {
			for _, item := range rPTR {
				records = append(records, item.NewNS(state.QName()))
			}
		}
	}

	return
}

func (r Redis) MX(ctx context.Context, zone string, state request.Request, previousRecords []dns.RR) (records []dns.RR, truncated bool, err error) {
	key := Key(state.Name(), r.KeyPrefix)
doSearch:
	val, err := r.get(ctx, key, state.Type())
	switch err {
	case nil:
		var rMX RecordMX
		err = json.Unmarshal([]byte(val), &rMX)
		if err != nil {
			return nil, false, err
		}
		for _, item := range rMX {
			records = append(records, item.NewMX(state.QName()))
		}

	case errKeyNotFound:
		stateNew := state.NewWithQuestion(state.Name(), dns.TypeCNAME)
		rCNAME, err := r.cname(ctx, zone, stateNew)
		if err != nil {
			if err == errKeyNotFound && !IsAnyKey(key) {
				key = AnyKey(key)
				goto doSearch
			}
			return nil, false, err
		}

		for _, item := range rCNAME {

			if len(previousRecords) > 7 {
				break
			}
			cnameRecode := item.NewCNAME(state.QName())
			if dnsutil.DuplicateCNAME(cnameRecode, previousRecords) {
				continue
			}

			if zone == "." || dns.IsSubDomain(zone, dns.Fqdn(item.Host)) {

				stateNew = state.NewWithQuestion(item.Host, state.QType())
				nextRecords, tc, err := r.MX(ctx, zone, stateNew, append(previousRecords, cnameRecode))
				if tc {
					truncated = true
				}

				if err == nil {
					if len(nextRecords) > 0 {
						records = append(records, cnameRecode)
						records = append(records, nextRecords...)
					}
					continue
				}

				if err != errKeyNotFound && zone != "." {
					continue
				}
			}

			m1, e1 := r.Lookup(ctx, state, cnameRecode.Target)
			if e1 != nil {
				continue
			}
			if m1.Truncated {
				truncated = true
			}
			records = append(records, cnameRecode)
			records = append(records, m1.Answer...)
		}

	default:
		if err != nil {
			return nil, false, err
		}
	}
	return records, truncated, nil
}

func (r Redis) SRV(ctx context.Context, zone string, state request.Request, previousRecords []dns.RR) (records []dns.RR, truncated bool, err error) {
	key := Key(state.Name(), r.KeyPrefix)
doSearch:
	val, err := r.get(ctx, key, state.Type())
	switch err {
	case nil:
		var rSRV RecordSRV
		err = json.Unmarshal([]byte(val), &rSRV)
		if err != nil {
			return nil, false, err
		}
		for _, item := range rSRV {
			records = append(records, item.NewSRV(state.QName()))
		}

	case errKeyNotFound:
		stateNew := state.NewWithQuestion(state.Name(), dns.TypeCNAME)
		rCNAME, err := r.cname(ctx, zone, stateNew)
		if err != nil {
			if err == errKeyNotFound && !IsAnyKey(key) {
				key = AnyKey(key)
				goto doSearch
			}
			return nil, false, err
		}

		for _, item := range rCNAME {

			if len(previousRecords) > 7 {
				break
			}
			cnameRecode := item.NewCNAME(state.QName())
			if dnsutil.DuplicateCNAME(cnameRecode, previousRecords) {
				continue
			}

			if zone == "." || dns.IsSubDomain(zone, dns.Fqdn(item.Host)) {

				stateNew = state.NewWithQuestion(item.Host, state.QType())
				nextRecords, tc, err := r.SRV(ctx, zone, stateNew, append(previousRecords, cnameRecode))
				if tc {
					truncated = true
				}

				if err == nil {
					if len(nextRecords) > 0 {
						records = append(records, cnameRecode)
						records = append(records, nextRecords...)
					}
					continue
				}

				if err != errKeyNotFound && zone != "." {
					continue
				}
			}

			m1, e1 := r.Lookup(ctx, state, cnameRecode.Target)
			if e1 != nil {
				continue
			}
			if m1.Truncated {
				truncated = true
			}
			records = append(records, cnameRecode)
			records = append(records, m1.Answer...)
		}

	default:
		if err != nil {
			return nil, false, err
		}
	}
	return records, truncated, nil
}

func (r Redis) CAA(ctx context.Context, zone string, state request.Request) (records []dns.RR, err error) {
	key := Key(state.Name(), r.KeyPrefix)

	val, err := r.get(ctx, key, state.Type())
	if err == nil {
		var rCAA RecordCAA
		err = json.Unmarshal([]byte(val), &rCAA)
		if err == nil {
			for _, item := range rCAA {
				records = append(records, item.NewCAA(state.QName()))
			}
		}
	}
	return
}

func (r Redis) SOA(ctx context.Context, zone string, state request.Request) ([]dns.RR, error) {
	key := Key(state.Name(), r.KeyPrefix)

	val, err := r.get(ctx, key, state.Type())

	switch err {
	case nil:
		var rSOA RecordSOA
		err = json.Unmarshal([]byte(val), &rSOA)
		if err != nil {
			return nil, err
		}
		return []dns.RR{ItemSOA(rSOA).NewSOA(state.QName())}, nil

	case errKeyNotFound:
		if zone != "." {
			header := dns.RR_Header{Name: zone, Rrtype: dns.TypeSOA, Ttl: r.MinTTL(state), Class: dns.ClassINET}

			Mbox := dnsutil.Join("hostmaster", zone)
			Ns := dnsutil.Join("ns.dns", zone)

			soa := &dns.SOA{Hdr: header,
				Mbox:    Mbox,
				Ns:      Ns,
				Serial:  uint32(time.Now().Unix()),
				Refresh: 7200,
				Retry:   1800,
				Expire:  86400,
				Minttl:  r.MinTTL(state),
			}
			return []dns.RR{soa}, nil
		}

	}
	return nil, err
}
