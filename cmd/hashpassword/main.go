
package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte("111"),
		bcrypt.DefaultCost,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(hash))
}
