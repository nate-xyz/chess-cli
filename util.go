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
	var rand_quote string = RandQuoteMap[rand_quote_int]
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

func timeFormat(time int64) string {
	if time == 0 {
		return "0"
	}
	ms := time % 1000
	time /= 1000
	sec := time % 60
	time /= 60
	min := time % 60
	hours := time / 60
	if hours == 0 && min == 0 && sec <= 10 {
		return fmt.Sprintf("%02d:%02d:%03d", min, sec, ms)
	} else if hours == 0 {
		return fmt.Sprintf("%02d:%02d", min, sec)
	}
	days := hours / 24
	hours = hours % 24
	if days == 0 {
		return fmt.Sprintf("%d Hours", hours)
	} else if hours == 0 {
		return fmt.Sprintf("%d Days", days)
	} else {
		return fmt.Sprintf("%d Days %d Hours", days, hours)
	}
}

func translateSelectedCell(row, col int, white bool) string {
	var rank string
	var file string
	if white {
		rank = fmt.Sprintf("%v", 8-(row-1))
		file = string(rune('a' + (col - 1)))
	} else {
		rank = fmt.Sprintf("%v", (row))
		file = string(rune('h' - (col - 1)))
	}
	return file + rank
}

func translateAlgtoCell(alg string, white bool) (r, c int) {
	file := alg[0]
	rank := alg[1]
	var row int
	var col int
	if white {
		row = -int(rank) + 57
		col = int(file) - 96
	} else {
		row = int(rank) - 48
		col = -int(file) + 105
	}
	return row, col
}
