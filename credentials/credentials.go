package credentials

import (
	"crypto/rand"
	"encoding/json"
	"log"

	"golang.org/x/crypto/scrypt"
)

//Credentials Structure that holds encrypted user credentials
type Credentials struct {
	Username string
	PHash    []byte
	Salt     []byte
}

//CheckPassword checks to see if the password matches the stored hash
func (c *Credentials) CheckPassword(password string) bool {
	dk, _ := scrypt.Key([]byte(password), c.Salt, 16384, 8, 1, 32)
	return compare(dk, c.PHash)
}

//CreateCredentials Creates new credentials with a new random salt
func CreateCredentials(username, password string) (c Credentials) {
	c.Salt = make([]byte, 32)
	c.Username = username
	_, err := rand.Read(c.Salt)
	if err != nil {
		log.Println("error:", err)
		c.Username = ""
		return
	}
	c.PHash, _ = scrypt.Key([]byte(password), c.Salt, 16384, 8, 1, 32)
	return
}

func compare(a, b []byte) bool {
	if &a == &b {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

//Json Converts the Credentials to JSON
func (c *Credentials) Json() (string, error) {
	j, e := json.Marshal(c)
	return string(j), e
}

//CreateFromJson Creates new credentials from a JSON representation
func CreateFromJson(j string) (Credentials, error) {
	c := Credentials{}
	err := json.Unmarshal([]byte(j), &c)
	return c, err
}
