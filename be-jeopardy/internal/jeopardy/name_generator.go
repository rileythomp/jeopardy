package jeopardy

import (
	_ "embed"
	"math/rand/v2"
	"strings"
)

//go:embed nouns.txt
var nounsList string
var nouns = strings.Split(nounsList, "\n")

//go:embed animals.txt
var animalsList string
var animals = strings.Split(animalsList, "\n")

//go:embed adjectives.txt
var adjectivesList string
var adjectives = strings.Split(adjectivesList, "\n")

func genGameName() string {
	return adjectives[rand.IntN(len(adjectives))] + "-" + nouns[rand.IntN(len(animals))]
}

func genBotName() string {
	return adjectives[rand.IntN(len(adjectives))] + "-" + animals[rand.IntN(len(animals))]
}

func genGameCode() string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"

	code := ""
	for i := 0; i < 3; i++ {
		code += string(letters[rand.IntN(len(letters))])
	}
	code += "-"
	for i := 0; i < 3; i++ {
		code += string(numbers[rand.IntN(len(numbers))])
	}

	return code
}
