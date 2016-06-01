package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
	"github.com/miekg/dns"
	)

func loadDummyRecords() {
	ar := new(dns.A)
	ar.Hdr = dns.RR_Header{Name: "testdns.dyndns.badpennies.co.uk.", Rrtype: dns.TypeA,
		Class: dns.ClassINET, Ttl: 3600}
	ar.A = net.ParseIP("192.168.1.1")
	rr, _ := dns.NewRR(ar.String())
	err := storeDNSRecord("testdns", rr)
	if err != nil {
		log.Println("failed to store ", err)
	}
	rr1, _ := getDNSRecord("test", rr.Header().Rrtype)
	log.Println("stored ", rr1)
}

func createNewDNSRecord(machine string, ip string) (rr dns.RR) {
	ipaddr := net.ParseIP(ip)
	if ipaddr.To4() == nil { // Its a V6 address
		ar := new(dns.AAAA)
		ar.Hdr = dns.RR_Header{Name: machine + "." + *rootDomain, Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: 300}
		ar.AAAA = ipaddr
		rr, _ = dns.NewRR(ar.String())
	} else {
		ar := new(dns.A)
		ar.Hdr = dns.RR_Header{Name: machine + "." + *rootDomain, Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: 300}
		ar.A = ipaddr
		rr, _ = dns.NewRR(ar.String())
	}

	return rr

}

func parseQuery(m *dns.Msg) {
	var rr dns.RR
	for _, q := range m.Question {
		l := dns.SplitDomainName(q.Name)
		if len(l) == len(dns.SplitDomainName(*rootDomain))+1 {
			if read_rr, e := getDNSRecord(l[0], q.Qtype); e == nil {
				rr = read_rr.(dns.RR)
				log.Println("Query Answered for ", rr.Header().Name)
				if rr.Header().Name == q.Name {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}
	if r.IsTsig() != nil {
		if w.TsigStatus() == nil {
			m.SetTsig(r.Extra[len(r.Extra)-1].(*dns.TSIG).Hdr.Name,
				dns.HmacMD5, 300, time.Now().Unix())
		} else {
			log.Println("Status ", w.TsigStatus().Error())
		}
	}

	w.WriteMsg(m)
}

func serve(name, secret string, port int) {
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}

	if name != "" {
		server.TsigSecret = map[string]string{name: secret}
	}

	err := server.ListenAndServe()
	defer server.Shutdown()

	if err != nil {
		log.Fatalf("Failed to setup the udp server: %s ", err.Error())
		fmt.Println("Failed to setup the udp server: %s ", err.Error())
	}
}
