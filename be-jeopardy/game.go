package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type GameState int

const (
	PreGame GameState = iota
	RecvPick
	RecvBuzz
	RecvWager
	RecvAns
	PostGame
)

type RoundState int

const (
	FirstRound RoundState = iota
	SecondRound
	FinalRound
)

const (
	numTopics    = 3
	numQuestions = 3
)

type (
	Game struct {
		State        GameState        `json:"state"`
		Round        RoundState       `json:"round"`
		Players      []*Player        `json:"players"`
		FirstRound   [numTopics]Topic `json:"firstRound"`
		SecondRound  [numTopics]Topic `json:"secondRound"`
		FinalRound   Question         `json:"finalRound"`
		CurQuestion  Question         `json:"curQuestion"`
		GuessedWrong []string         `json:"guessedWrong"`
		LastPicker   string           `json:"lastPicker"`
		FinalWagers  int              `json:"finalWagers"`
		FinalAnswers int              `json:"finalAnswers"`
		Passes       int              `json:"passes"`
	}

	Player struct {
		Id         string `json:"id"`
		Name       string `json:"name"`
		Score      int    `json:"score"`
		CanPick    bool   `json:"canPick"`
		CanBuzz    bool   `json:"canBuzz"`
		CanAnswer  bool   `json:"canAnswer"`
		CanWager   bool   `json:"canWager"`
		FinalWager int    `json:"finalWager"`

		conn *websocket.Conn
	}

	Topic struct {
		Title     string                 `json:"title"`
		Questions [numQuestions]Question `json:"questions"`
	}

	Question struct {
		Question    string `json:"question"`
		Answer      string `json:"answer"`
		Value       int    `json:"value"`
		CanChoose   bool   `json:"canChoose"`
		DailyDouble bool   `json:"dailyDouble"`
	}
)

func NewGame() *Game {
	return &Game{
		State:   PreGame,
		Players: []*Player{},
	}
}

func (g *Game) setState(state GameState, id string) {
	switch state {
	case RecvPick:
		for _, player := range g.Players {
			player.CanPick = player.Id == id
			player.CanBuzz = false
			player.CanAnswer = false
			player.CanWager = false
		}
	case RecvBuzz:
		for _, player := range g.Players {
			player.CanPick = false
			player.CanBuzz = player.canBuzz(g.GuessedWrong)
			player.CanAnswer = false
			player.CanWager = false
		}
	case RecvAns:
		for _, player := range g.Players {
			player.CanPick = false
			player.CanBuzz = false
			player.CanAnswer = player.Id == id
			if g.Round == FinalRound {
				player.CanAnswer = player.CanWager
			}
			player.CanWager = false
		}
	case RecvWager:
		for _, player := range g.Players {
			player.CanPick = false
			player.CanBuzz = false
			player.CanAnswer = false
			player.CanWager = player.Id == id
			if g.Round == FinalRound {
				player.CanWager = player.Score > 0
			}
		}
	default:
		for _, player := range g.Players {
			player.CanPick = false
			player.CanBuzz = false
			player.CanAnswer = false
			player.CanWager = false
		}
	}
	g.State = state
}

func (g *Game) handlePick(playerId string, topicIdx, valIdx int) error {
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("Player not found")
	}
	if !player.CanPick {
		return fmt.Errorf("Player cannot pick")
	}
	if topicIdx < 0 || valIdx < 0 || topicIdx >= numTopics || valIdx >= numQuestions {
		return fmt.Errorf("invalid question pick")
	}

	curRound := g.FirstRound
	if g.Round == SecondRound {
		curRound = g.SecondRound
	}
	curQuestion := curRound[topicIdx].Questions[valIdx]
	if !curQuestion.CanChoose {
		return fmt.Errorf("Question cannot be chosen")
	}
	g.LastPicker = player.Id
	g.CurQuestion = curQuestion
	var resp Response
	if curQuestion.DailyDouble {
		g.setState(RecvWager, player.Id)
		resp = Response{
			Code:    200,
			Message: "Daily Double",
			Game:    g,
		}
	} else {
		g.setState(RecvBuzz, "")
		resp = Response{
			Code:    200,
			Message: "New Question",
			Game:    g,
		}
	}
	if err := g.messageAllPlayers(resp); err != nil {
		return err
	}
	return nil
}

func (g *Game) handleBuzz(playerId string, isPass bool) error {
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("Player not found")
	}
	if !player.CanBuzz {
		return fmt.Errorf("Player cannot buzz")
	}
	if isPass {
		g.Passes++
	}
	var resp Response
	if g.Passes == len(g.Players) {
		g.disableQuestion()
		g.Passes = 0
		g.setState(RecvPick, g.LastPicker)
		resp = Response{
			Code:    200,
			Message: "Question unanswered",
			Game:    g,
		}
	} else {
		g.setState(RecvAns, player.Id)
		resp = Response{
			Code:    200,
			Message: "Player buzzed",
			Game:    g,
		}
	}
	if err := g.messageAllPlayers(resp); err != nil {
		return err
	}
	return nil
}

func (g *Game) handleAnswer(playerId, answer string) error {
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("Player not found")
	}
	if !player.CanAnswer {
		return fmt.Errorf("Player cannot answer")
	}
	isCorrect := g.CurQuestion.checkAnswer(answer)
	player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
	var resp Response
	if g.Round == FinalRound {
		g.FinalAnswers++
		if !g.roundEnded() {
			log.Printf("received answer from %s: %s\n", player.Name, answer)
			return nil
		}
		g.setState(PostGame, "")
		resp = Response{
			Code:    200,
			Message: "Final round ended",
			Game:    g,
		}
	} else {
		if !isCorrect {
			g.GuessedWrong = append(g.GuessedWrong, player.Id)
		}
		if isCorrect || g.CurQuestion.DailyDouble || len(g.GuessedWrong) == len(g.Players) {
			g.disableQuestion()
		}
		roundOver := g.roundEnded()
		if roundOver && g.Round == FirstRound {
			g.Round = SecondRound
			g.GuessedWrong = []string{}
			g.setState(RecvPick, g.lowestPlayer())
			resp = Response{
				Code:    200,
				Message: "First round ended",
				Game:    g,
			}
		} else if roundOver && g.Round == SecondRound {
			g.Round = FinalRound
			g.CurQuestion = g.FinalRound
			g.setState(RecvWager, "")
			resp = Response{
				Code:    200,
				Message: "Second round ended",
				Game:    g,
			}
		} else if len(g.GuessedWrong) == len(g.Players) {
			g.GuessedWrong = []string{}
			g.setState(RecvPick, g.LastPicker)
			resp = Response{
				Code:    200,
				Message: "All players guessed wrong",
				Game:    g,
			}
		} else if isCorrect || g.CurQuestion.DailyDouble {
			g.GuessedWrong = []string{}
			g.setState(RecvPick, playerId)
			resp = Response{
				Code:    200,
				Message: "Question is complete",
				Game:    g,
			}
		} else {
			g.setState(RecvBuzz, "")
			resp = Response{
				Code:    200,
				Message: "Player answered incorrectly",
				Game:    g,
			}
		}
	}
	if err := g.messageAllPlayers(resp); err != nil {
		return err
	}
	return nil
}

func (g *Game) handleWager(playerId string, wager int) error {
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("Player not found")
	}
	if !player.CanWager {
		return fmt.Errorf("Player cannot wager")
	}
	if !g.validWager(wager, player.Score) {
		return fmt.Errorf("invalid wager, must be between 5 and %d", max(player.Score, 1000))
	}
	var resp Response
	if g.Round == FinalRound {
		player.FinalWager = wager
		g.FinalWagers++
		if g.FinalWagers != g.numFinalWagers() {
			resp = Response{
				Code:    200,
				Message: "Player wagered",
				Game:    g,
			}
			if err := player.conn.WriteJSON(resp); err != nil {
				return err
			}
			return nil
		}
		g.setState(RecvAns, "")
		resp = Response{
			Code:    200,
			Message: "All wagers received",
			Game:    g,
		}
	} else {
		g.CurQuestion.Value = wager
		g.setState(RecvAns, player.Id)
		resp = Response{
			Code:    200,
			Message: "Player wagered",
			Game:    g,
		}
	}
	if err := g.messageAllPlayers(resp); err != nil {
		return err
	}
	return nil
}

func (p *Player) updateScore(val int, isCorrect bool, round RoundState) {
	if round == FinalRound {
		val = p.FinalWager
	}
	if !isCorrect {
		val *= -1
	}
	p.Score += val
}

func (g *Game) roundEnded() bool {
	if g.Round == FinalRound {
		log.Printf("final answers: %d, num final answers: %d\n", g.FinalAnswers, g.numFinalAnswers())
		return g.FinalAnswers == g.numFinalAnswers()
	}
	curRound := g.FirstRound
	if g.Round == SecondRound {
		curRound = g.SecondRound
	}
	for _, topic := range curRound {
		for _, question := range topic.Questions {
			if question.CanChoose {
				return false
			}
		}
	}
	return true
}

func (g *Game) messageAllPlayers(resp Response) error {
	for _, player := range game.Players {
		if player.conn != nil {
			resp.CurPlayer = player
			if err := player.conn.WriteJSON(resp); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Game) addPlayer(name string) string {
	player := NewPlayer(name)
	g.Players = append(g.Players, player)
	return player.Id
}

func (g *Game) numPlayersReady() int {
	playersReady := 0
	for i := range g.Players {
		if g.Players[i].conn != nil {
			playersReady++
		}
	}
	return playersReady
}

func (g *Game) readyToPlay() bool {
	return g.numPlayersReady() == 3
}

func (g *Game) getPlayerById(id string) *Player {
	for _, player := range g.Players {
		if player.Id == id {
			return player
		}
	}
	return nil
}

func (g *Game) hasStarted() bool {
	return g.State != PreGame
}

func (g *Game) setQuestions() error {
	g.FirstRound = [numTopics]Topic{
		{
			Title: "World Capitals",
			Questions: [numQuestions]Question{
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
					Question:  "This city is the capital of France",
					Answer:    "Paris",
					Value:     600,
					CanChoose: true,
				},
				// {
				// 	Question:    "This city is the capital of Germany",
				// 	Answer:      "Berlin",
				// 	Value:       800,
				// 	CanChoose:   true,
				// 	DailyDouble: true,
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
			Questions: [numQuestions]Question{
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
			Questions: [numQuestions]Question{
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
		// 	Questions: [numQuestions]Question{
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
		// 	Questions: [numQuestions]Question{
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
		// 	Questions: [numQuestions]Question{
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

	g.SecondRound = [numTopics]Topic{
		{
			Title: "Movie Trivia",
			Questions: [numQuestions]Question{
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
			Questions: [numQuestions]Question{
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
			Questions: [numQuestions]Question{
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
					Question:  "This is the largest organ in the human body",
					Answer:    "The skin",
					Value:     1200,
					CanChoose: true,
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
		// 	Questions: [numQuestions]Question{
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
		// 	Questions: [numQuestions]Question{
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
		// 	Questions: [numQuestions]Question{
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
	g.FinalRound = Question{
		Question: "An MLB team got this name in 1902 after some of its players defected to a new crosstown rival, leaving young replacements",
		Answer:   "Chicago Cubs",
	}
	return nil
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
	g.CurQuestion = Question{}
}

func (g *Game) validWager(wager, score int) bool {
	minWager := 5
	if g.Round == FinalRound {
		minWager = 0
	}
	roundMax := 0
	if g.Round == FirstRound {
		roundMax = 1000
	} else if g.Round == SecondRound {
		roundMax = 2000
	}
	return wager >= minWager && wager <= max(score, roundMax)
}

func (g *Game) numFinalWagers() int {
	numWagers := 0
	for _, player := range g.Players {
		if player.CanWager {
			numWagers++
		}
	}
	return numWagers
}

func (g *Game) numFinalAnswers() int {
	numAnswers := 0
	for _, player := range g.Players {
		if player.CanAnswer {
			numAnswers++
		}
	}
	return numAnswers
}

func NewPlayer(name string) *Player {
	return &Player{
		Id:        uuid.New().String(),
		Name:      name,
		Score:     0,
		CanPick:   false,
		CanBuzz:   false,
		CanAnswer: false,
	}
}

func (p *Player) canBuzz(guessedWrong []string) bool {
	for _, id := range guessedWrong {
		if id == p.Id {
			return false
		}
	}
	return true
}

func (q *Question) checkAnswer(ans string) bool {
	ans = strings.ToLower(ans)
	corrAns := strings.ToLower(q.Answer)
	if strings.Contains(ans, corrAns) || strings.Contains(corrAns, ans) {
		return true
	}
	return levenshtein.ComputeDistance(ans, corrAns) < 3
}

func (g *Game) lowestPlayer() string {
	lowest := g.Players[0]
	for _, player := range g.Players {
		if player.Score < lowest.Score {
			lowest = player
		}
	}
	return lowest.Id
}

func (q *Question) equal(q0 Question) bool {
	return q.Question == q0.Question && q.Answer == q0.Answer
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
