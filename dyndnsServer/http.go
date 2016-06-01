package main

import (
	"encoding/base64"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"172.16.100.1/droker/godyndns/credentials"
	"github.com/miekg/dns"
)

func processHTTPSetRequest(w http.ResponseWriter, r *http.Request) {
	host := getHostIP(r)
	username, _ := getUsernamePassword(r)
	err := storeDNSRecord(username, createNewDNSRecord(username, host))
	if err != nil {
		io.WriteString(w, "Unable to set DNS record for "+username)
	} else {
		io.WriteString(w, "DNS record for "+username+" set to "+host)
		log.Println("SET: DNS record for " + username + " set to " + host)
	}

}

func processHTTPGetRequest(w http.ResponseWriter, r *http.Request) {
	//host := getHostIP(r)
	username, _ := getUsernamePassword(r)
	rrV4, err := getDNSRecord(username, dns.TypeA)
	if err != nil {
		io.WriteString(w, "No V4 DNS record exists for "+username)
	} else {
		io.WriteString(w, username+" is at "+rrV4.String())
		log.Println("GET DNS A record for " + username + " is at " + rrV4.String())
	}
	rrV6, err := getDNSRecord(username, dns.TypeAAAA)
	if err != nil {
		io.WriteString(w, "No V6 DNS record exists for "+username)
	} else {
		io.WriteString(w, username+" is at "+rrV6.String())
		log.Println("GET DNS A record for " + username + " is at " + rrV4.String())
	}

}

func processHTTPNewUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if strings.Join(r.Form["username"], "") == "" || strings.Join(r.Form["password"], "") == "" {
		http.Error(w, "Internal Server", 500)
		return
	}
	c := credentials.CreateCredentials(strings.Join(r.Form["username"], ""), strings.Join(r.Form["password"], ""))
	err := storeUserRecord(c)
	if err != nil {
		http.Error(w, "Problem storing User", 500)
	} else {
		io.WriteString(w, "username added "+c.Username)
		log.Println("New user added " + c.Username)
	}
}

func login(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			t, _ := template.ParseFiles("login.gtpl")
			t.Execute(w, nil)
		} else {
			h.ServeHTTP(w, r)
		}
	}
}

func processHTTPRequest(w http.ResponseWriter, r *http.Request) {
	host := getHostIP(r)
	io.WriteString(w, "Hello world! from "+host)
}

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}
		if pair[0] == *adminUser {
			http.Error(w, "Not authorized", 401)
			return
		} else {
			c, err := getUserRecord(pair[0])
			if err != nil {

			}
			if c.CheckPassword(pair[1]) == false {
				http.Error(w, "Not authorized", 401)
				return
			}
		}

		h.ServeHTTP(w, r)
	}
}

func adminAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}
		if pair[0] == *adminUser {
			if pair[1] != *adminPassword {
				http.Error(w, "Not authorized", 401)
				return
			}
		}
		h.ServeHTTP(w, r)
	}
}

func getUsernamePassword(r *http.Request) (username, password string) {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	b, _ := base64.StdEncoding.DecodeString(s[1])
	pair := strings.SplitN(string(b), ":", 2)
	return pair[0], pair[1]
}

func getHostIP(r *http.Request) string {
	s := r.RemoteAddr
	host, _, _ := net.SplitHostPort(s)
	return host
}
