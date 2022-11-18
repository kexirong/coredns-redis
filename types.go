package redis

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

type RecordSOA ItemSOA
type RecordA []ItemIP
type RecordAAAA []ItemIP
type RecordTXT []ItemText
type RecordCNANE []ItemHost
type RecordNS []ItemHost
type RecordPTR []ItemHost
type RecordMX []ItemMX
type RecordSRV []ItemSRV
type RecordCAA []ItemCAA

type ItemIP struct {
	TTL uint32 `json:"ttl,omitempty"`
	IP  net.IP `json:"ip"`
}

func (i ItemIP) NewA(name string) *dns.A {
	return &dns.A{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: i.TTL}, A: i.IP}
}

func (i ItemIP) NewAAAA(name string) *dns.AAAA {
	return &dns.AAAA{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: i.TTL}, AAAA: i.IP}
}

type ItemText struct {
	TTL  uint32 `json:"ttl,omitempty"`
	Text string `json:"text"`
}

func (i ItemText) NewTXT(name string) *dns.TXT {
	return &dns.TXT{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: i.TTL}, Txt: Split255(i.Text)}
}

type ItemHost struct {
	TTL  uint32 `json:"ttl,omitempty"`
	Host string `json:"host"`
}

func (i ItemHost) NewCNAME(name string) *dns.CNAME {
	return &dns.CNAME{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: i.TTL}, Target: dns.Fqdn(i.Host)}
}

func (i ItemHost) NewNS(name string) *dns.NS {
	return &dns.NS{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: i.TTL}, Ns: dns.Fqdn(i.Host)}
}

func (i ItemHost) NewPTR(name string) *dns.PTR {
	return &dns.PTR{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: i.TTL}, Ptr: dns.Fqdn(i.Host)}
}

type ItemMX struct {
	ItemHost
	Preference uint16 `json:"preference"`
}

func (i ItemMX) NewMX(name string) *dns.MX {
	return &dns.MX{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: i.TTL}, Mx: dns.Fqdn(i.Host), Preference: i.Preference}
}

type ItemSRV struct {
	TTL      uint32 `json:"ttl,omitempty"`
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
	Target   string `json:"target"`
}

func (i ItemSRV) NewSRV(name string) *dns.SRV {
	return &dns.SRV{
		Hdr:      dns.RR_Header{Name: name, Rrtype: dns.TypeSRV, Class: dns.ClassINET, Ttl: i.TTL},
		Priority: i.Priority,
		Port:     i.Port,
		Weight:   i.Weight,
		Target:   dns.Fqdn(i.Target),
	}
}

type ItemSOA struct {
	NS      string `json:"ns"`
	Mbox    string `json:"Mbox"`
	Refresh uint32 `json:"refresh"`
	Retry   uint32 `json:"retry"`
	Expire  uint32 `json:"expire"`
	MinTTL  uint32 `json:"minTTL"`
}

func (i ItemSOA) NewSOA(name string) *dns.SOA {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: i.MinTTL},
		Mbox:    dns.Fqdn(i.Mbox),
		Ns:      dns.Fqdn(i.NS),
		Serial:  uint32(time.Now().Unix()),
		Refresh: i.Refresh,
		Retry:   i.Retry,
		Expire:  i.Expire,
		Minttl:  i.MinTTL,
	}
}

type ItemCAA struct {
	TTL   uint32 `json:"ttl,omitempty"`
	Flag  uint8  `json:"flag"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

func (i ItemCAA) NewCAA(name string) *dns.CAA {
	return &dns.CAA{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeCAA, Class: dns.ClassINET}, Flag: i.Flag, Tag: i.Tag, Value: i.Value}
}
