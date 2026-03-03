package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	stored := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lHuy"
	err := bcrypt.CompareHashAndPassword([]byte(stored), []byte("admin123"))
	fmt.Println("match err:", err)

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	fmt.Println("new hash:", string(hash))
}
