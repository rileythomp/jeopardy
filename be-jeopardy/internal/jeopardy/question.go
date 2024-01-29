package jeopardy

import (
	"strings"

	"github.com/agnivade/levenshtein"
)

type (
	Topic struct {
		Title     string     `json:"title"`
		Questions []Question `json:"questions"`
	}

	Question struct {
		Question    string `json:"question"`
		Answer      string `json:"answer"`
		Value       int    `json:"value"`
		CanChoose   bool   `json:"canChoose"`
		DailyDouble bool   `json:"dailyDouble"`
	}
)

const (
	numTopics    = 3
	numQuestions = 3
)

func (q *Question) checkAnswer(ans string) bool {
	ans = strings.ToLower(ans)
	corrAns := strings.ToLower(q.Answer)
	if len(ans) < 5 {
		return ans == corrAns
	} else if len(corrAns) < 7 {
		return levenshtein.ComputeDistance(ans, corrAns) < 2
	} else if len(corrAns) < 9 {
		return levenshtein.ComputeDistance(ans, corrAns) < 3
	} else if len(corrAns) < 11 {
		return levenshtein.ComputeDistance(ans, corrAns) < 4
	} else if len(corrAns) < 13 {
		return levenshtein.ComputeDistance(ans, corrAns) < 5
	} else if len(corrAns) < 15 {
		return levenshtein.ComputeDistance(ans, corrAns) < 6
	}
	return levenshtein.ComputeDistance(ans, corrAns) < 7
}

func (q *Question) equal(q0 Question) bool {
	return q.Question == q0.Question && q.Answer == q0.Answer
}

func (g *Game) setQuestions() error {
	g.FirstRound = []Topic{
		{
			Title: "World Capitals",
			Questions: []Question{
				{
					Question:  "This city is the capital of the United States",
					Answer:    "Washington, D.C.",
					Value:     200,
					CanChoose: true,
				},
				{
					Question:  "This city is the capital of the United Kingdom",
					Answer:    "London",
					Value:     400,
					CanChoose: true,
				},
				{
					Question:    "This city is the capital of France",
					Answer:      "Paris",
					Value:       600,
					CanChoose:   true,
					DailyDouble: false,
				},
				// {
				// 	Question:    "This city is the capital of Germany",
				// 	Answer:      "Berlin",
				// 	Value:       800,
				// 	CanChoose:   true,
				// },
				// {
				// 	Question:  "This city is the capital of Russia",
				// 	Answer:    "Moscow",
				// 	Value:     1000,
				// 	CanChoose: true,
				// },
			},
		},
		{
			Title: "State Capitals",
			Questions: []Question{
				{
					Question:  "This city is the capital of California",
					Answer:    "Sacramento",
					Value:     200,
					CanChoose: true,
				},
				{
					Question:  "This city is the capital of Texas",
					Answer:    "Austin",
					Value:     400,
					CanChoose: true,
				},
				{
					Question:  "This city is the capital of New York",
					Answer:    "Albany",
					Value:     600,
					CanChoose: true,
				},
				// {
				// 	Question:  "This city is the capital of Florida",
				// 	Answer:    "Tallahassee",
				// 	Value:     800,
				// 	CanChoose: true,
				// },
				// {
				// 	Question:  "This city is the capital of Washington",
				// 	Answer:    "Olympia",
				// 	Value:     1000,
				// 	CanChoose: true,
				// },
			},
		},
		{
			Title: "Provincial Capitals",
			Questions: []Question{
				{
					Question:  "This city is the capital of British Columbia",
					Answer:    "Victoria",
					Value:     200,
					CanChoose: true,
				},
				{
					Question:  "This city is the capital of Alberta",
					Answer:    "Edmonton",
					Value:     400,
					CanChoose: true,
				},
				{
					Question:  "This city is the capital of Saskatchewan",
					Answer:    "Regina",
					Value:     600,
					CanChoose: true,
				},
				// {
				// 	Question:  "This city is the capital of Manitoba",
				// 	Answer:    "Winnipeg",
				// 	Value:     800,
				// 	CanChoose: true,
				// },
				// {
				// 	Question:  "This city is the capital of Ontario",
				// 	Answer:    "Toronto",
				// 	Value:     1000,
				// 	CanChoose: true,
				// },
			},
		},
		// {
		// 	Title: "Sports Trivia",
		// 	Questions: []Question{
		// 		{
		// 			Question:  "This team won the 2019 NBA Finals",
		// 			Answer:    "Toronto Raptors",
		// 			Value:     200,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This team won the 2019 Stanley Cup",
		// 			Answer:    "St. Louis Blues",
		// 			Value:     400,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This team won the 2019 World Series",
		// 			Answer:    "Washington Nationals",
		// 			Value:     600,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This team won the 2019 Super Bowl",
		// 			Answer:    "New England Patriots",
		// 			Value:     800,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This team won the 2019 MLS Cup",
		// 			Answer:    "Seattle Sounders",
		// 			Value:     1000,
		// 			CanChoose: true,
		// 		},
		// 	},
		// },
		// {
		// 	Title: "Music Trivia",
		// 	Questions: []Question{
		// 		{
		// 			Question:  "This artist won the 2019 Grammy for Album of the Year",
		// 			Answer:    "Kacey Musgraves",
		// 			Value:     200,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This artist won the 2019 Grammy for Record of the Year",
		// 			Answer:    "Childish Gambino",
		// 			Value:     400,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This artist won the 2019 Grammy for Song of the Year",
		// 			Answer:    "Donald Glover",
		// 			Value:     600,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This artist won the 2019 Grammy for Best New Artist",
		// 			Answer:    "Dua Lipa",
		// 			Value:     800,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This artist won the 2019 Grammy for Best Rap Album",
		// 			Answer:    "Igor",
		// 			Value:     1000,
		// 			CanChoose: true,
		// 		},
		// 	},
		// },
		// {
		// 	Title: "Geography Trivia",
		// 	Questions: []Question{
		// 		{
		// 			Question:  "This is the largest country in the world",
		// 			Answer:    "Russia",
		// 			Value:     200,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the largest country in Africa",
		// 			Answer:    "Algeria",
		// 			Value:     400,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the largest country in South America",
		// 			Answer:    "Brazil",
		// 			Value:     600,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the largest country in Asia",
		// 			Answer:    "China",
		// 			Value:     800,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the largest country in Europe, excluding Russia",
		// 			Answer:    "Ukraine",
		// 			Value:     1000,
		// 			CanChoose: true,
		// 		},
		// 	},
		// },
	}

	g.SecondRound = []Topic{
		{
			Title: "Movie Trivia",
			Questions: []Question{
				{
					Question:  "This movie won the 2019 Oscar for Best Picture",
					Answer:    "Green Book",
					Value:     400,
					CanChoose: true,
				},
				{
					Question:  "This movie won the 2019 Oscar for Best Animated Feature",
					Answer:    "Spider-Man: Into the Spider-Verse",
					Value:     800,
					CanChoose: true,
				},
				{
					Question:  "This movie won the 2019 Oscar for Best Actor",
					Answer:    "Rami Malek",
					Value:     1200,
					CanChoose: true,
				},
				// {
				// 	Question:  "This movie won the 2019 Oscar for Best Actress",
				// 	Answer:    "Olivia Colman",
				// 	Value:     1600,
				// 	CanChoose: true,
				// },
				// {
				// 	Question:  "This movie won the 2019 Oscar for Best Director",
				// 	Answer:    "Alfonso Cuarón",
				// 	Value:     2000,
				// 	CanChoose: true,
				// },
			},
		},
		{
			Title: "TV Trivia",
			Questions: []Question{
				{
					Question:  "This show won the 2019 Emmy for Best Drama Series",
					Answer:    "Game of Thrones",
					Value:     400,
					CanChoose: true,
				},
				{
					Question:  "This show won the 2019 Emmy for Best Comedy Series",
					Answer:    "Fleabag",
					Value:     800,
					CanChoose: true,
				},
				{
					Question:  "This actor won the 2019 Emmy for Best Actor in a Drama Series",
					Answer:    "Billy Porter",
					Value:     1200,
					CanChoose: true,
				},
				// {
				// 	Question:  "This actress won the 2019 Emmy for Best Actress in a Drama Series",
				// 	Answer:    "Jodie Comer",
				// 	Value:     1600,
				// 	CanChoose: true,
				// },
				// {
				// 	Question:  "This actress won the 2019 Emmy for Best Actress in a Comedy Series",
				// 	Answer:    "Phoebe Waller-Bridge",
				// 	Value:     2000,
				// 	CanChoose: true,
				// },
			},
		},
		{
			Title: "Science Trivia",
			Questions: []Question{
				{
					Question:  "This is the largest planet in the solar system",
					Answer:    "Jupiter",
					Value:     400,
					CanChoose: true,
				},
				{
					Question:  "This is the smallest planet in the solar system",
					Answer:    "Mercury",
					Value:     800,
					CanChoose: true,
				},
				{
					Question:    "This is the largest organ in the human body",
					Answer:      "The skin",
					Value:       1200,
					CanChoose:   true,
					DailyDouble: false,
				},
				// {
				// 	Question:  "This is the largest bone in the human body",
				// 	Answer:    "The femur",
				// 	Value:     1600,
				// 	CanChoose: true,
				// },
				// {
				// 	Question:  "This is the world's largest animal",
				// 	Answer:    "The Antarctic blue whale",
				// 	Value:     2000,
				// 	CanChoose: true,
				// },
			},
		},
		// {
		// 	Title: "History Trivia",
		// 	Questions: []Question{
		// 		{
		// 			Question:  "This is the year that WWII ended",
		// 			Answer:    "1945",
		// 			Value:     400,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the year that the Berlin Wall fell",
		// 			Answer:    "1989",
		// 			Value:     800,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the year that the Titanic sank",
		// 			Answer:    "1912",
		// 			Value:     1200,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the year that the Magna Carta was signed",
		// 			Answer:    "1215",
		// 			Value:     1600,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the year that the Declaration of Independence was signed",
		// 			Answer:    "1776",
		// 			Value:     2000,
		// 			CanChoose: true,
		// 		},
		// 	},
		// },
		// {
		// 	Title: "Math Trivia",
		// 	Questions: []Question{
		// 		{
		// 			Question:  "This is the longest side of a right triangle",
		// 			Answer:    "Hypotenuse",
		// 			Value:     400,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the number of degrees in a circle",
		// 			Answer:    "360",
		// 			Value:     800,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the number of degrees in a right angle",
		// 			Answer:    "90",
		// 			Value:     1200,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the number of degrees in a straight angle",
		// 			Answer:    "180",
		// 			Value:     1600,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "This is the number of degrees in a triangle",
		// 			Answer:    "180",
		// 			Value:     2000,
		// 			CanChoose: true,
		// 		},
		// 	},
		// },
		// {
		// 	Title: "Business Trivia",
		// 	Questions: []Question{
		// 		{
		// 			Question:  "This 3-letter memorandum of debt is a strong but not legally binding promise to pay",
		// 			Answer:    "I.O.U.",
		// 			Value:     400,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "In 2007 Forbes reported that this TV personality was \"America's sole black female billionaire\"",
		// 			Answer:    "Oprah Winfrey",
		// 			Value:     800,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "26 billion merger in 2016, this business website might keep nagging Microsoft to update its resume",
		// 			Answer:    "LinkedIn",
		// 			Value:     1200,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "A passage from the book of Hosea was the inspiration Israel's first Minister of Transportation had for naming this airline	El",
		// 			Answer:    "El Al",
		// 			Value:     1600,
		// 			CanChoose: true,
		// 		},
		// 		{
		// 			Question:  "Corn is traded at this sort of exchange as well as, of course, frozen concentrated orange juice",
		// 			Answer:    "Commodity",
		// 			Value:     2000,
		// 			CanChoose: true,
		// 		},
		// 	},
		// },
	}
	g.FinalQuestion = Question{
		Question: "An MLB team got this name in 1902 after some of its players defected to a new crosstown rival, leaving young replacements",
		Answer:   "Chicago Cubs",
	}
	return nil
}

func (g *Game) firstAvailableQuestion() (int, int) {
	curRound := g.FirstRound
	if g.Round == SecondRound {
		curRound = g.SecondRound
	}
	for valIdx := 0; valIdx < numQuestions; valIdx++ {
		for topicIdx := 0; topicIdx < numTopics; topicIdx++ {
			if curRound[topicIdx].Questions[valIdx].CanChoose {
				return topicIdx, valIdx
			}
		}

	}
	return -1, -1
}

func (g *Game) disableQuestion() {
	for i, topic := range g.FirstRound {
		for j, q := range topic.Questions {
			if q.equal(g.CurQuestion) {
				g.FirstRound[i].Questions[j].CanChoose = false
			}
		}
	}
	for i, topic := range g.SecondRound {
		for j, q := range topic.Questions {
			if q.equal(g.CurQuestion) {
				g.SecondRound[i].Questions[j].CanChoose = false
			}
		}
	}
}
