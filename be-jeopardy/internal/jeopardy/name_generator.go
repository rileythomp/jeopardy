package jeopardy

import (
	_ "embed"
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

func genGameCode() string {
	return adjectives[rng.Intn(len(adjectives))] + "-" + nouns[rng.Intn(len(animals))]
}

func genBotCode() string {
	return adjectives[rng.Intn(len(adjectives))] + "-" + animals[rng.Intn(len(animals))]
}
