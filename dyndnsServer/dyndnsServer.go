package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"github.com/boltdb/bolt"
	"github.com/miekg/dns"
	//"172.16.100.1/droker/godyndns/credentials"
)

var (
	tsig          *string
	dbPath        *string
	port          *int
	httpPort      *int
	zoneDB        *bolt.DB
	bindIP        *string
	logFile       *string
	pidFile       *string
	rootDomain    *string
	adminUser     *string
	adminPassword *string
	bucket        = "zones"
)

func main() {
	// Parse flags
	logFile = flag.String("logfile", "", "path to log file")
	port = flag.Int("port", 53, "server port")
	httpPort = flag.Int("httpport", 8000, "server port")
	bindIP = flag.String("bind", "", "server port")
	dbPath = flag.String("dbPath", "./dyndns.db", "location where db will be stored")
	pidFile = flag.String("pid", "./go-dyndns.pid", "pid file location")
	rootDomain = flag.String("rootdomain", "dyndns.example.co.uk.", "root domain of dns server")
	adminUser = flag.String("adminuser", "", "Admin Username")
	adminPassword = flag.String("adminPass", "", "Admin password")

	flag.Parse()
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}
	if *adminUser == "" || *adminPassword == "" {
		log.Fatalf("Must set admin username and password")
	}
	if err := openDatabase(); err != nil {
		log.Fatalln(err.Error())
	}
	dns.HandleFunc(".", handleDnsRequest)
	http.HandleFunc("/", processHTTPRequest)
	http.HandleFunc("/set", basicAuth(processHTTPSetRequest))
	http.HandleFunc("/get", basicAuth(processHTTPGetRequest))
	http.HandleFunc("/add", adminAuth(login(processHTTPNewUser)))
	go serve("", "secret", *port)

	err := http.ListenAndServeTLS(*bindIP +":" + strconv.Itoa(*httpPort), "server.crt", "server.key", nil) //(":8000", nil)
	if err != nil {
		log.Fatalf("Failed to setup the HTTPs server: %s ", err.Error())
		fmt.Println("Failed to setup the HTTPs server: %s ", err.Error())
	}
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
endless:
	for {
		select {
		case s := <-sig:
			log.Printf("Signal (%d) received, stoppingn ", s)
			break endless
		}
	}
}
