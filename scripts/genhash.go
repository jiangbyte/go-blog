//go:build ignore

package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	b, err := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
