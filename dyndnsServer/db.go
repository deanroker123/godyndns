package main

import (
	"errors"
	"log"
	"strings"
	"time"
	"172.16.100.1/droker/godyndns/credentials"
	"github.com/boltdb/bolt"
	"github.com/miekg/dns"
)

func createBucket(bucket string) (err error) {
	err = zoneDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			e := errors.New("Create bucket:  " + bucket)
			log.Println(e.Error())

			return e
		}

		return nil
	})

	return err
}

func openDatabase() error {
	var err error
	zoneDB, err = bolt.Open(*dbPath, 0600,
		&bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return err
	}

	err = zoneDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket + "A"))
		if err != nil {
			e := errors.New("Create bucket:  " + bucket + "A")
			return e
		}
		_, err = tx.CreateBucketIfNotExists([]byte(bucket + "AAAA"))
		if err != nil {
			e := errors.New("Create bucket:  " + bucket + "AAAA")
			return e
		}
		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			e := errors.New("Create bucket: users")
			return e
		}
		return nil
	})
	return err
}

func storeDNSRecord(u string, rr dns.RR) (err error) {

	err = zoneDB.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		switch rr.Header().Rrtype {
		case dns.TypeA:
			b = tx.Bucket([]byte(bucket + "A"))
		case dns.TypeAAAA:
			b = tx.Bucket([]byte(bucket + "AAAA"))
		default:
			e := errors.New("Unsupported DNS Record Type")
			log.Println(e.Error())
			return e
		}

		err := b.Put([]byte(u), []byte(rr.String()))

		if err != nil {
			e := errors.New("Store record failed:  " + rr.String())
			log.Println(e.Error())
			return e
		}
		return nil
	})
	return err
}

func getDNSRecord(u string, r uint16) (rr dns.RR, err error) {
	//key, _ := getKey(domain, rtype)
	var v []byte
	err = zoneDB.View(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		switch r {
		case dns.TypeA:
			b = tx.Bucket([]byte(bucket + "A"))
		case dns.TypeAAAA:
			b = tx.Bucket([]byte(bucket + "AAAA"))
		default:
			e := errors.New("Unsupported DNS Record Type")
			log.Println(e.Error())
			return e
		}
		v = b.Get([]byte(u))
		if string(v) == "" {
			e := errors.New("Record not found, machine:  " + u)
			log.Println(e.Error())
			return e
		}
		return nil
	})
	if err == nil {
		rr, err = dns.NewRR(string(v))
	}
	return rr, err
}

func storeUserRecord(c credentials.Credentials) (err error) {

	err = zoneDB.Update(func(tx *bolt.Tx) error {
		jc, _ := c.Json()
		b := tx.Bucket([]byte("users"))
		err := b.Put([]byte(strings.ToLower(c.Username)), []byte(jc))

		if err != nil {
			e := errors.New("Store record failed:  " + c.Username)
			log.Println(e.Error())
			return e
		}
		return nil
	})
	return err
}

func getUserRecord(u string) (credentials.Credentials, error) {
	//key, _ := getKey(domain, rtype)
	var v []byte
	err := zoneDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		v = b.Get([]byte(strings.ToLower(u)))

		if string(v) == "" {
			e := errors.New("Record not found, user:  " + u)
			log.Println(e.Error())

		}
		return nil
	})
	var cred credentials.Credentials
	if err == nil {
		cred, err = credentials.CreateFromJson(string(v))
	}
	return cred, err
}
