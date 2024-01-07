package jeopardy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	RecvAnsConfirmation
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
		cancelRecvAns     map[string]context.CancelFunc
		cancelRecvBuzz    context.CancelFunc
		State             GameState        `json:"state"`
		Round             RoundState       `json:"round"`
		Players           []*Player        `json:"players"`
		FirstRound        [numTopics]Topic `json:"firstRound"`
		SecondRound       [numTopics]Topic `json:"secondRound"`
		FinalQuestion     Question         `json:"finalQuestion"`
		CurQuestion       Question         `json:"curQuestion"`
		GuessedWrong      []string         `json:"guessedWrong"`
		LastPicker        string           `json:"lastPicker"`
		NumFinalWagers    int              `json:"numFinalWagers"`
		FinalWagersRecvd  int              `json:"finalWagers"`
		FinalAnswersRecvd int              `json:"finalAnswers"`
		Passes            int              `json:"passes"`
		LastAnswer        string           `json:"lastAnswer"`
		AnsCorrectness    bool             `json:"ansCorrectness"`
		Confirmations     int              `json:"confirmations"`
		Challenges        int              `json:"challenges"`
		LastAnswerer      *Player          `json:"lastAnswerer"`
	}

	Player struct {
		Id              string          `json:"id"`
		Name            string          `json:"name"`
		Score           int             `json:"score"`
		CanPick         bool            `json:"canPick"`
		CanBuzz         bool            `json:"canBuzz"`
		CanAnswer       bool            `json:"canAnswer"`
		CanWager        bool            `json:"canWager"`
		CanConfirmAns   bool            `json:"canConfirmAns"`
		FinalWager      int             `json:"finalWager"`
		FinalAnswer     string          `json:"finalAnswer"`
		FinalCorrect    bool            `json:"finalCorrect"`
		FinalProtestors map[string]bool `json:"finalProtestors"`

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
		State:         PreGame,
		Players:       []*Player{},
		cancelRecvAns: map[string]context.CancelFunc{},
	}
}

func (g *Game) AddPlayer(name string) (string, error) {
	if g.State != PreGame {
		return "", fmt.Errorf("game already in progress")
	}
	if len(g.Players) > 2 {
		return "", fmt.Errorf("game is full")
	}
	player := NewPlayer(name)
	g.Players = append(g.Players, player)
	return player.Id, nil
}

func (g *Game) SetPlayerConnection(playerId string, conn *websocket.Conn) error {
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	player.conn = conn
	msg := "Waiting for more players"
	if g.readyToPlay() {
		if err := g.startGame(); err != nil {
			return fmt.Errorf("error starting game: %w", err)
		}
		msg = "We are ready to play"
	}
	if err := g.messageAllPlayers(msg); err != nil {
		return fmt.Errorf("error sending message to players: %w", err)
	}
	return nil
}

func (g *Game) readyToPlay() bool {
	playersReady := 0
	for i := range g.Players {
		if g.Players[i].conn != nil {
			playersReady++
		}
	}
	return playersReady == 3
}

func (g *Game) startGame() error {
	if err := g.setQuestions(); err != nil {
		return err
	}
	g.setState(RecvPick, g.Players[0].Id)
	// for i := range g.Players {
	// 	// random score between 1000 and 5000
	// 	g.Players[i].Score = (rand.Intn(5) + 1) * 1000
	// }
	// g.startFinalRound()
	return nil
}

func (g *Game) HandleRequest(playerId string, msg []byte) error {
	var err error
	switch g.State {
	case RecvPick:
		var pickReq PickRequest
		if err := json.Unmarshal(msg, &pickReq); err != nil {
			return fmt.Errorf("failed to parse pick request: %w", err)
		}
		err = g.handlePick(playerId, pickReq.TopicIdx, pickReq.ValIdx)
	case RecvBuzz:
		var buzzReq BuzzRequest
		if err := json.Unmarshal(msg, &buzzReq); err != nil {
			return fmt.Errorf("failed to parse buzz request: %w", err)
		}
		err = g.handleBuzz(playerId, buzzReq.IsPass)
	case RecvAns:
		var ansReq AnswerRequest
		if err := json.Unmarshal(msg, &ansReq); err != nil {
			return fmt.Errorf("failed to parse answer request: %w", err)
		}
		err = g.handleAnswer(playerId, ansReq.Answer)
	case RecvAnsConfirmation:
		var confAnsReq ConfirmAnsRequest
		if err := json.Unmarshal(msg, &confAnsReq); err != nil {
			return fmt.Errorf("failed to parse confirm answer request: %w", err)
		}
		err = g.handleAnsConfirmation(playerId, confAnsReq.Confirm)
	case RecvWager:
		var wagerReq WagerRequest
		if err := json.Unmarshal(msg, &wagerReq); err != nil {
			return fmt.Errorf("failed to parse wager request: %w", err)
		}
		err = g.handleWager(playerId, wagerReq.Wager)
	case PostGame:
		var protestReq ProtestRequest
		if err := json.Unmarshal(msg, &protestReq); err != nil {
			return fmt.Errorf("failed to parse protest request: %w", err)
		}
		err = g.handleProtest(protestReq.ProtestFor, playerId)
	default:
		err = fmt.Errorf("invalid game state")
	}
	return err
}

func (g *Game) handleProtest(protestFor, protestBy string) error {
	protestForPlayer := g.getPlayerById(protestFor)
	protestByPlayer := g.getPlayerById(protestBy)
	if protestForPlayer == nil || protestByPlayer == nil {
		return fmt.Errorf("player not found")
	}
	if _, ok := protestForPlayer.FinalProtestors[protestByPlayer.Id]; ok {
		return nil
	}
	protestForPlayer.FinalProtestors[protestByPlayer.Id] = true
	if len(protestForPlayer.FinalProtestors) == 2 {
		if protestForPlayer.FinalCorrect {
			protestForPlayer.Score -= 2 * protestForPlayer.FinalWager
		} else {
			protestForPlayer.Score += 2 * protestForPlayer.FinalWager
		}
		g.setState(PostGame, "")
		g.messageAllPlayers("final jeopardy result changed")
	} else {
		protestByPlayer.conn.WriteJSON(Response{
			Code:      200,
			Message:   "You protested for " + protestForPlayer.Name,
			Game:      g,
			CurPlayer: protestByPlayer,
		})
	}
	return nil
}

func (g *Game) handlePick(playerId string, topicIdx, valIdx int) error {
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if !player.CanPick {
		return fmt.Errorf("player cannot pick")
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
		return fmt.Errorf("question cannot be chosen")
	}
	g.LastPicker = player.Id
	g.CurQuestion = curQuestion
	var msg string
	if curQuestion.DailyDouble {
		g.setState(RecvWager, player.Id)
		msg = "Daily Double"
	} else {
		g.setState(RecvBuzz, "")
		msg = "New Question"
	}
	return g.messageAllPlayers(msg)
}

func (g *Game) skipQuestion() error {
	var msg string
	g.disableQuestion()
	roundOver := g.roundEnded()
	if roundOver && g.Round == FirstRound {
		g.startSecondRound()
		msg = "First round ended"
	} else if roundOver && g.Round == SecondRound {
		g.startFinalRound()
		msg = "Second round ended"
	} else {
		g.resetGuesses()
		g.setState(RecvPick, g.LastPicker)
		msg = "Question unanswered"
	}
	return g.messageAllPlayers(msg)
}

func (g *Game) handleBuzz(playerId string, isPass bool) error {
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if !player.CanBuzz {
		return fmt.Errorf("player cannot buzz")
	}
	if isPass {
		g.Passes++
		player.CanBuzz = false
		if g.Passes+len(g.GuessedWrong) != len(g.Players) {
			return player.conn.WriteJSON(Response{
				Code:      200,
				Message:   "You passed",
				Game:      g,
				CurPlayer: player,
			})
		}
		g.cancelRecvBuzz()
		return g.skipQuestion()
	}
	g.cancelRecvBuzz()
	g.setState(RecvAns, player.Id)
	return g.messageAllPlayers("Player buzzed")
}

func (g *Game) handleAnswer(playerId, answer string) error {
	cancelRecvAns := g.cancelRecvAns[playerId]
	cancelRecvAns()
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if !player.CanAnswer {
		return fmt.Errorf("player cannot answer")
	}
	isCorrect := g.CurQuestion.checkAnswer(answer)
	var msg string
	if g.Round == FinalRound {
		player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
		g.FinalAnswersRecvd++
		player.CanAnswer = false
		player.FinalAnswer = answer
		player.FinalCorrect = isCorrect
		if !g.roundEnded() {
			return player.conn.WriteJSON(Response{
				Code:      200,
				Message:   "You answered",
				Game:      g,
				CurPlayer: player,
			})
		}
		g.setState(PostGame, "")
		msg = "Final round ended"
	} else {
		isCorrect := g.CurQuestion.checkAnswer(answer)
		g.AnsCorrectness = isCorrect
		g.LastAnswer = answer
		g.LastAnswerer = player
		g.setState(RecvAnsConfirmation, "")
		msg = "Player answered"
	}
	return g.messageAllPlayers(msg)
}

func (g *Game) handleAnsConfirmation(playerId string, confirm bool) error {
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if !player.CanConfirmAns {
		return fmt.Errorf("player cannot confirm")
	}

	player.CanConfirmAns = false
	if confirm {
		g.Confirmations++
	} else {
		g.Challenges++
	}
	if g.Confirmations != 2 && g.Challenges != 2 {
		return player.conn.WriteJSON(Response{
			Code:      200,
			Message:   "You confirmed",
			Game:      g,
			CurPlayer: player,
		})
	}

	isCorrect := (g.Confirmations == 2 && g.AnsCorrectness) || (g.Challenges == 2 && !g.AnsCorrectness)
	g.LastAnswerer.updateScore(g.CurQuestion.Value, isCorrect, g.Round)

	var msg string
	if g.Round == FinalRound {
		panic("should not be here")
		// g.FinalAnswersRecvd++
		// player.CanAnswer = false
		// if !g.roundEnded() {
		// 	return player.conn.WriteJSON(Response{
		// 		Code:      200,
		// 		Message:   "You answered",
		// 		Game:      g,
		// 		CurPlayer: player,
		// 	})
		// }
		// g.setState(PostGame, "")
		// msg = "Final round ended 22222"
	} else {
		if !isCorrect {
			g.GuessedWrong = append(g.GuessedWrong, g.LastAnswerer.Id)
		}
		if isCorrect || g.CurQuestion.DailyDouble || g.Passes+len(g.GuessedWrong) == len(g.Players) {
			g.disableQuestion()
		}
		roundOver := g.roundEnded()
		if roundOver && g.Round == FirstRound {
			g.startSecondRound()
			msg = "First round ended"
		} else if roundOver && g.Round == SecondRound {
			g.startFinalRound()
			msg = "Second round ended"
		} else if g.Passes+len(g.GuessedWrong) == len(g.Players) {
			g.resetGuesses()
			g.setState(RecvPick, g.LastPicker)
			msg = "All players guessed wrong"
		} else if isCorrect || g.CurQuestion.DailyDouble {
			g.resetGuesses()
			g.setState(RecvPick, g.LastAnswerer.Id)
			msg = "Question is complete"
		} else {
			g.Confirmations = 0
			g.Challenges = 0
			g.setState(RecvBuzz, "")
			msg = "Player answered incorrectly"
		}
	}
	return g.messageAllPlayers(msg)
}

func (g *Game) handleWager(playerId string, wager int) error {
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if !player.CanWager {
		return fmt.Errorf("player cannot wager")
	}
	if min, max, ok := g.validWager(wager, player.Score); !ok {
		return player.conn.WriteJSON(Response{
			Code:      400,
			Message:   fmt.Sprintf("invalid wager, must be between %d and %d", min, max),
			Game:      g,
			CurPlayer: player,
		})
	}
	var msg string
	if g.Round == FinalRound {
		player.FinalWager = wager
		player.CanWager = false
		g.FinalWagersRecvd++
		if g.FinalWagersRecvd != g.NumFinalWagers {
			return player.conn.WriteJSON(Response{
				Code:      200,
				Message:   "You wagered",
				Game:      g,
				CurPlayer: player,
			})
		}
		g.setState(RecvAns, "")
		msg = "All wagers received"
	} else {
		// daily double
		g.CurQuestion.Value = wager
		g.setState(RecvAns, player.Id)
		msg = "Player wagered"
	}
	return g.messageAllPlayers(msg)
}

func (g *Game) getPlayerById(id string) *Player {
	for _, player := range g.Players {
		if player.Id == id {
			return player
		}
	}
	return nil
}

func (g *Game) setState(state GameState, id string) {
	switch state {
	case RecvPick:
		for _, player := range g.Players {
			player.CanPick = player.Id == id
			player.CanBuzz = false
			player.CanAnswer = false
			player.CanWager = false
			player.CanConfirmAns = false
		}
	case RecvBuzz:
		for _, player := range g.Players {
			player.CanPick = false
			player.CanBuzz = player.canBuzz(g.GuessedWrong)
			player.CanAnswer = false
			player.CanWager = false
			player.CanConfirmAns = false
		}
		recvBuzzCtx, cancelRecvBuzz := context.WithCancel(context.Background())
		g.cancelRecvBuzz = cancelRecvBuzz
		go func(recvBuzzCtx context.Context) {
			timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 7*time.Second)
			defer timeoutCancel()
			select {
			case <-recvBuzzCtx.Done():
				fmt.Println("Cancelling buzz in timeout")
				return
			case <-timeoutCtx.Done():
				fmt.Println("7 seconds elapsed with no buzz, skipping question")
				err := g.skipQuestion()
				if err != nil {
					panic(err)
				}
				return
			}
		}(recvBuzzCtx)
	case RecvAns:
		for _, player := range g.Players {
			player.CanPick = false
			player.CanBuzz = false
			player.CanAnswer = player.Id == id
			if g.Round == FinalRound {
				player.CanAnswer = player.Score > 0
			}
			player.CanWager = false
			player.CanConfirmAns = false
		}
		for _, player := range g.Players {
			if !player.CanAnswer {
				continue
			}
			recvAnsCtx, cancelRecvAns := context.WithCancel(context.Background())
			g.cancelRecvAns[player.Id] = cancelRecvAns
			go func(recvAnsCtx context.Context, playerId string) {
				answerTimeout := 10 * time.Second
				if g.CurQuestion.DailyDouble {
					answerTimeout = 20 * time.Second
				} else if g.Round == FinalRound {
					answerTimeout = 30 * time.Second
				}
				timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), answerTimeout)
				defer timeoutCancel()
				select {
				case <-recvAnsCtx.Done():
					fmt.Println("Cancelling answer in timeout")
					return
				case <-timeoutCtx.Done():
					fmt.Printf("%d seconds elapsed with no answer, skipping question\n", answerTimeout/time.Second)
					err := g.handleAnswer(playerId, "answer-timeout")
					if err != nil {
						panic(err)
					}
					return
				}
			}(recvAnsCtx, player.Id)
		}
	case RecvAnsConfirmation:
		for _, player := range g.Players {
			player.CanPick = false
			player.CanBuzz = false
			player.CanAnswer = false
			player.CanWager = false
			player.CanConfirmAns = true
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
			player.CanConfirmAns = false
		}
	default:
		for _, player := range g.Players {
			player.CanPick = false
			player.CanBuzz = false
			player.CanAnswer = false
			player.CanWager = false
			player.CanConfirmAns = false
		}
	}
	g.State = state
}

func (g *Game) messageAllPlayers(msg string) error {
	for _, player := range g.Players {
		if player.conn != nil {
			if err := player.conn.WriteJSON(Response{
				Code:      200,
				Message:   msg,
				Game:      g,
				CurPlayer: player,
			}); err != nil {
				// TODO: HANDLE ERROR SYNCHRONIZATION
				return err
			}
		}
	}
	return nil
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
					Question:    "This city is the capital of France",
					Answer:      "Paris",
					Value:       600,
					CanChoose:   true,
					DailyDouble: true,
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
					Question:    "This is the largest organ in the human body",
					Answer:      "The skin",
					Value:       1200,
					CanChoose:   true,
					DailyDouble: true,
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
	g.FinalQuestion = Question{
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
}

func (g *Game) validWager(wager, score int) (int, int, bool) {
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
	return minWager, max(score, roundMax), wager >= minWager && wager <= max(score, roundMax)
}

func (g *Game) roundEnded() bool {
	if g.Round == FinalRound {
		return g.FinalAnswersRecvd == g.NumFinalWagers
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

func (g *Game) resetGuesses() {
	g.GuessedWrong = []string{}
	g.Passes = 0
	g.Confirmations = 0
	g.Challenges = 0
}

func (g *Game) startSecondRound() {
	g.Round = SecondRound
	g.resetGuesses()
	g.setState(RecvPick, g.lowestPlayer())
}

func (g *Game) startFinalRound() {
	g.Round = FinalRound
	g.resetGuesses()
	g.CurQuestion = g.FinalQuestion
	g.NumFinalWagers = g.numFinalWagers()
	if g.NumFinalWagers < 2 {
		g.setState(PostGame, "")
	} else {
		g.setState(RecvWager, "")
	}
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

func (g *Game) numFinalWagers() int {
	numWagers := 0
	for _, player := range g.Players {
		if player.Score > 0 {
			numWagers++
		}
	}
	return numWagers
}

func NewPlayer(name string) *Player {
	return &Player{
		Id:              uuid.New().String(),
		Name:            name,
		Score:           0,
		CanPick:         false,
		CanBuzz:         false,
		CanAnswer:       false,
		CanWager:        false,
		CanConfirmAns:   false,
		FinalProtestors: map[string]bool{},
	}
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
