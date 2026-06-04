// hashgen: OAuth 어드민 비밀번호의 bcrypt hash 를 생성하는 일회용 도구.
// 사용: go run ./cmd/hashgen <password>
package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/hashgen <password>")
		os.Exit(1)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), 12)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bcrypt 생성 실패:", err)
		os.Exit(1)
	}
	fmt.Println(string(hash))
}
