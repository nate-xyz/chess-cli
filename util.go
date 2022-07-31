package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func GetRandomQuote() string {
	rand.Seed(time.Now().UnixNano())
	minrand := 0
	maxrand := 56
	rand_quote_int := rand.Intn(maxrand-minrand+1) + minrand
	var rand_quote string = rand_quote_map[rand_quote_int]
	return rand_quote
}

//writes to a local text file
func WriteLocal(title string, payload string) {
	f, err := os.Create(fmt.Sprintf("%s.txt", title))

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(payload)

	if err2 != nil {
		log.Fatal(err2)
	}
}
