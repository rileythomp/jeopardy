package jeopardy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log"
)

type (
	Game struct {
		State             GameState  `json:"state"`
		Players           []*Player  `json:"players"`
		Round             RoundState `json:"round"`
		FirstRound        []Topic    `json:"firstRound"`
		SecondRound       []Topic    `json:"secondRound"`
		FinalQuestion     Question   `json:"finalQuestion"`
		CurQuestion       Question   `json:"curQuestion"`
		GuessedWrong      []string   `json:"guessedWrong"`
		LastPicker        string     `json:"lastPicker"`
		NumFinalWagers    int        `json:"numFinalWagers"`
		FinalWagersRecvd  int        `json:"finalWagers"`
		FinalAnswersRecvd int        `json:"finalAnswers"`
		Passes            int        `json:"passes"`
		LastAnswer        string     `json:"lastAnswer"`
		AnsCorrectness    bool       `json:"ansCorrectness"`
		Confirmations     int        `json:"confirmations"`
		Challenges        int        `json:"challenges"`
		LastAnswerer      *Player    `json:"lastAnswerer"`

		cancelRecvAns             map[string]context.CancelFunc
		cancelRecvWager           map[string]context.CancelFunc
		cancelRecvAnsConfirmation context.CancelFunc
		cancelRecvBuzz            context.CancelFunc
		cancelRecvPick            context.CancelFunc
	}

	Request struct {
		PickRequest
		BuzzRequest
		AnswerRequest
		ConfirmAnsRequest
		WagerRequest
		ProtestRequest
	}

	PickRequest struct {
		TopicIdx int `json:"topicIdx"`
		ValIdx   int `json:"valIdx"`
	}

	BuzzRequest struct {
		IsPass bool `json:"isPass"`
	}

	AnswerRequest struct {
		Answer string `json:"answer"`
	}

	ConfirmAnsRequest struct {
		Confirm bool `json:"confirm"`
	}

	WagerRequest struct {
		Wager int `json:"wager"`
	}

	ProtestRequest struct {
		ProtestFor string `json:"protestFor"`
	}

	Response struct {
		Code      int     `json:"code"`
		Token     string  `json:"token,omitempty"`
		Message   string  `json:"message"`
		Game      *Game   `json:"game,omitempty"`
		CurPlayer *Player `json:"curPlayer,omitempty"`
	}
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
	numPlayers = 3

	pickQuestionTimeout       = 9 * time.Second
	buzzInTimeout             = 12 * time.Second
	defaultAnsTimeout         = 10 * time.Second
	dailyDoubleAnsTimeout     = 10 * time.Second
	finalJeopardyAnsTimeout   = 10 * time.Second
	confirmAnsTimeout         = 10 * time.Second
	dailyDoubleWagerTimeout   = 10 * time.Second
	finalJeopardyWagerTimeout = 10 * time.Second
)

var (
	games       = []*Game{}
	playerGames = map[string]*Game{}
)

func NewGame() *Game {
	return &Game{
		State:           PreGame,
		Players:         []*Player{},
		Round:           FirstRound,
		cancelRecvAns:   map[string]context.CancelFunc{},
		cancelRecvWager: map[string]context.CancelFunc{},
	}
}

func GetGame(playerId string) *Game {
	return playerGames[playerId]
}

func TerminateGames() {
	for _, game := range games {
		game.terminateGame()
	}
	playerGames = map[string]*Game{}
	games = []*Game{}
}

func JoinGame(playerName string) (*Game, string, error) {
	var game = NewGame()
	for _, g := range games {
		if len(g.Players) < 3 {
			game = g
		}
	}
	playerId, err := game.addPlayer(playerName)
	if err != nil {
		log.Printf("Error adding player to game: %s", err.Error())
		return nil, "", err
	}
	playerGames[playerId] = game
	if len(game.Players) == 1 {
		games = append(games, game)
	}
	return game, playerId, nil
}

func PlayGame(playerId string, conn SafeConn) error {
	game := GetGame(playerId)
	if game == nil {
		return fmt.Errorf("no game found for player id: %s", playerId)
	}

	player := game.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("no player found for player id")
	}
	player.conn = conn

	msg := "Waiting for more players"
	if game.readyToPlay() {
		if err := game.startGame(); err != nil {
			return err
		}
		msg = "We are ready to play"
	}
	if err := game.messageAllPlayers(msg); err != nil {
		return err
	}

	go func() {
		// TODO: USE A CHANNEL TO WAIT ON A MESSAGE OR TO END THE GAME
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message from WebSocket: %s\n", err.Error())
				player.closeConnWithMsg("Error reading message from WebSocket")
				return
			}
			err = game.HandleRequest(playerId, msg)
			if err != nil {
				log.Printf("Error handling request: %s\n", err.Error())
				player.closeConnWithMsg(err.Error())
				return
			}
		}
	}()

	return nil
}

func (g *Game) HandleRequest(playerId string, msg []byte) error {
	var req Request
	var err error
	if err := json.Unmarshal(msg, &req); err != nil {
		log.Printf("Error parsing request: %s\n", err.Error())
		return fmt.Errorf("error parsing request")
	}
	switch g.State {
	case RecvPick:
		err = g.handlePick(playerId, req.TopicIdx, req.ValIdx)
	case RecvBuzz:
		err = g.handleBuzz(playerId, req.IsPass)
	case RecvAns:
		err = g.handleAnswer(playerId, req.Answer)
	case RecvAnsConfirmation:
		err = g.handleAnsConfirmation(playerId, req.Confirm)
	case RecvWager:
		err = g.handleWager(playerId, req.Wager)
	case PostGame:
		err = g.handleProtest(playerId, req.ProtestFor)
	default:
		err = fmt.Errorf("invalid game state")
	}
	return err
}

func (g *Game) handlePick(playerId string, topicIdx, valIdx int) error {
	g.cancelRecvPick()
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
			return player.sendMessage(Response{
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
		return g.handleFinalRoundAns(player, isCorrect, answer)
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

func (g *Game) handleAnswerTimeout(playerId string) error {
	cancelRecvAns := g.cancelRecvAns[playerId]
	cancelRecvAns()
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if !player.CanAnswer {
		return fmt.Errorf("player cannot answer")
	}
	isCorrect := false
	if g.Round == FinalRound {
		return g.handleFinalRoundAns(player, isCorrect, "answer-timeout")
	}
	return g.nextQuestion(player, isCorrect)
}

func (g *Game) handleFinalRoundAns(player *Player, isCorrect bool, answer string) error {
	player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
	g.FinalAnswersRecvd++
	player.CanAnswer = false
	player.FinalAnswer = answer
	player.FinalCorrect = isCorrect
	if !g.roundEnded() {
		return player.sendMessage(Response{
			Code:      200,
			Message:   "You answered",
			Game:      g,
			CurPlayer: player,
		})
	}
	g.setState(PostGame, "")
	return g.messageAllPlayers("Final round ended")
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
		return player.sendMessage(Response{
			Code:      200,
			Message:   "You confirmed",
			Game:      g,
			CurPlayer: player,
		})
	}
	if g.Round == FinalRound {
		return fmt.Errorf("should not be handling answer confirmation in final round")
	}
	g.cancelRecvAnsConfirmation()
	isCorrect := (g.AnsCorrectness && g.Confirmations == 2) || (!g.AnsCorrectness && g.Challenges == 2)
	return g.nextQuestion(g.LastAnswerer, isCorrect)
}

func (g *Game) handleWager(playerId string, wager int) error {
	cancelRecvWager := g.cancelRecvWager[playerId]
	cancelRecvWager()
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if !player.CanWager {
		return fmt.Errorf("player cannot wager")
	}
	if min, max, ok := g.validWager(wager, player.Score); !ok {
		return player.sendMessage(Response{
			Code:      http.StatusBadRequest,
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
			return player.sendMessage(Response{
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

func (g *Game) handleProtest(playerId, protestFor string) error {
	protestForPlayer := g.getPlayerById(protestFor)
	protestByPlayer := g.getPlayerById(playerId)
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
		return g.messageAllPlayers("final jeopardy result changed")
	}
	return protestByPlayer.sendMessage(Response{
		Code:      200,
		Message:   "You protested for " + protestForPlayer.Name,
		Game:      g,
		CurPlayer: protestByPlayer,
	})
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
		recvPickCtx, cancelRecvPick := context.WithCancel(context.Background())
		g.cancelRecvPick = cancelRecvPick
		go func(recvPickCtx context.Context) {
			timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), pickQuestionTimeout)
			defer timeoutCancel()
			select {
			case <-recvPickCtx.Done():
				fmt.Println("Cancelling pick question timeout")
				return
			case <-timeoutCtx.Done():
				fmt.Printf("%d seconds elapsed with no pick, automatically choosing question\n", pickQuestionTimeout/time.Second)
				topicIdx, valIdx := g.firstAvailableQuestion()
				err := g.handlePick(id, topicIdx, valIdx)
				if err != nil {
					log.Printf("Unexpected error picking question after timeout: %s\n", err)
					g.terminateGame()
				}
				return
			}
		}(recvPickCtx)
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
			timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), buzzInTimeout)
			defer timeoutCancel()
			select {
			case <-recvBuzzCtx.Done():
				fmt.Println("Cancelling buzz in timeout")
				return
			case <-timeoutCtx.Done():
				fmt.Printf("%d seconds elapsed with no buzz, skipping question", buzzInTimeout/time.Second)
				err := g.skipQuestion()
				if err != nil {
					log.Printf("Unexpected error skipping question after timeout: %s\n", err)
					g.terminateGame()
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
				answerTimeout := defaultAnsTimeout
				if g.CurQuestion.DailyDouble {
					answerTimeout = dailyDoubleAnsTimeout
				} else if g.Round == FinalRound {
					answerTimeout = finalJeopardyAnsTimeout
				}
				timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), answerTimeout)
				defer timeoutCancel()
				select {
				case <-recvAnsCtx.Done():
					fmt.Println("Cancelling answer in timeout")
					return
				case <-timeoutCtx.Done():
					fmt.Printf("%d seconds elapsed with no answer, skipping question\n", answerTimeout/time.Second)
					err := g.handleAnswerTimeout(playerId)
					if err != nil {
						log.Printf("Unexpected error skipping answer after timeout: %s\n", err)
						g.terminateGame()
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
		recvAnsConfirmationCtx, cancelRecvAnsConfirmation := context.WithCancel(context.Background())
		g.cancelRecvAnsConfirmation = cancelRecvAnsConfirmation
		go func(recvAnsConfirmationCtx context.Context) {
			timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), confirmAnsTimeout)
			defer timeoutCancel()
			select {
			case <-recvAnsConfirmationCtx.Done():
				fmt.Println("Cancelling answer confirmation in timeout")
				return
			case <-timeoutCtx.Done():
				fmt.Printf("%d seconds elapsed with no answer confirmation, automatically confirming\n", confirmAnsTimeout/time.Second)
				err := g.nextQuestion(g.LastAnswerer, g.AnsCorrectness)
				if err != nil {
					log.Printf("Unexpected error skipping answer confirmation after timeout: %s\n", err)
					g.terminateGame()
				}
				return
			}
		}(recvAnsConfirmationCtx)
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
		for _, player := range g.Players {
			if !player.CanWager {
				continue
			}
			recvWagerCtx, cancelRecvWager := context.WithCancel(context.Background())
			g.cancelRecvWager[player.Id] = cancelRecvWager
			go func(recvWagerCtx context.Context, playerId string) {
				wagerTimeout := dailyDoubleWagerTimeout
				if g.Round == FinalRound {
					wagerTimeout = finalJeopardyWagerTimeout
				}
				timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), wagerTimeout)
				defer timeoutCancel()
				select {
				case <-recvWagerCtx.Done():
					fmt.Println("Cancelling wager in timeout")
					return
				case <-timeoutCtx.Done():
					fmt.Printf("%d seconds elapsed with no wager, wagering 0 automatically\n", wagerTimeout/time.Second)
					wager := 5
					if g.Round == FinalRound {
						wager = 0
					}
					err := g.handleWager(playerId, wager)
					if err != nil {
						log.Printf("Unexpected error skipping wager after timeout: %s\n", err)
						g.terminateGame()
					}
					return
				}
			}(recvWagerCtx, player.Id)
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

func (g *Game) addPlayer(name string) (string, error) {
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

func (g *Game) nextQuestion(player *Player, isCorrect bool) error {
	player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
	if !isCorrect {
		g.GuessedWrong = append(g.GuessedWrong, player.Id)
	}
	if isCorrect || g.CurQuestion.DailyDouble || g.Passes+len(g.GuessedWrong) == len(g.Players) {
		g.disableQuestion()
	}
	var msg string
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
		g.setState(RecvPick, player.Id)
		msg = "Question is complete"
	} else {
		g.Confirmations = 0
		g.Challenges = 0
		g.setState(RecvBuzz, "")
		msg = "Player answered incorrectly"
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

func (g *Game) terminateGame() {
	// TODO: HANDLE ERROR SYNCHRONIZATION
	log.Print("Terminating game\n")
	if g.cancelRecvPick != nil {
		g.cancelRecvPick()
	}
	if g.cancelRecvBuzz != nil {
		g.cancelRecvBuzz()
	}
	if g.cancelRecvAnsConfirmation != nil {
		g.cancelRecvAnsConfirmation()
	}
	for _, player := range g.Players {
		cancelRecvAns, ok := g.cancelRecvAns[player.Id]
		if ok {
			cancelRecvAns()
		}
		cancelRecvWager, ok := g.cancelRecvWager[player.Id]
		if ok {
			cancelRecvWager()
		}
		_ = player.closeConnection()
	}
}

func (g *Game) resetGuesses() {
	g.GuessedWrong = []string{}
	g.Passes = 0
	g.Confirmations = 0
	g.Challenges = 0
}

func (g *Game) getPlayerById(id string) *Player {
	for _, player := range g.Players {
		if player.Id == id {
			return player
		}
	}
	return nil
}

func (g *Game) messageAllPlayers(msg string) error {
	for _, player := range g.Players {
		if err := player.sendMessage(Response{
			Code:      200,
			Message:   msg,
			Game:      g,
			CurPlayer: player,
		}); err != nil {
			// TODO: HANDLE ERROR SYNCHRONIZATION
			return err
		}
	}
	return nil
}

func (g *Game) readyToPlay() bool {
	ready := 0
	for _, player := range g.Players {
		if player.conn != nil {
			ready++
		}
	}
	return ready == numPlayers
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
