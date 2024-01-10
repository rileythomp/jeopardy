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
		Round             RoundState `json:"round"`
		FirstRound        []Topic    `json:"firstRound"`
		SecondRound       []Topic    `json:"secondRound"`
		FinalQuestion     Question   `json:"finalQuestion"`
		CurQuestion       Question   `json:"curQuestion"`
		Players           []*Player  `json:"players"`
		LastAnswerer      *Player    `json:"lastAnswerer"`
		LastAnswer        string     `json:"lastAnswer"`
		LastPicker        string     `json:"lastPicker"`
		GuessedWrong      []string   `json:"guessedWrong"`
		AnsCorrectness    bool       `json:"ansCorrectness"`
		NumFinalWagers    int        `json:"numFinalWagers"`
		FinalWagersRecvd  int        `json:"finalWagers"`
		FinalAnswersRecvd int        `json:"finalAnswers"`
		Passes            int        `json:"passes"`
		Confirmations     int        `json:"confirmations"`
		Challenges        int        `json:"challenges"`

		cancelPickTimeout         context.CancelFunc
		cancelBuzzTimeout         context.CancelFunc
		cancelConfirmationTimeout context.CancelFunc
		cancelAnswerTimeout       map[string]context.CancelFunc
		cancelWagerTimeout        map[string]context.CancelFunc

		cancelSendingPings chan bool
		sendPingsTicker    *time.Ticker
	}

	Message struct {
		PickMessage
		BuzzMessage
		AnswerMessage
		ConfirmAnsMessage
		WagerMessage
		ProtestMessage
	}

	PickMessage struct {
		TopicIdx int `json:"topicIdx"`
		ValIdx   int `json:"valIdx"`
	}

	BuzzMessage struct {
		IsPass bool `json:"isPass"`
	}

	AnswerMessage struct {
		Answer string `json:"answer"`
	}

	ConfirmAnsMessage struct {
		Confirm bool `json:"confirm"`
	}

	WagerMessage struct {
		Wager int `json:"wager"`
	}

	ProtestMessage struct {
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
		State:               PreGame,
		Players:             []*Player{},
		Round:               FirstRound,
		cancelAnswerTimeout: map[string]context.CancelFunc{},
		cancelWagerTimeout:  map[string]context.CancelFunc{},
	}
}

func GetGames() []*Game {
	return games
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
		game.cancelSendingPings = make(chan bool)
		game.sendPingsTicker = time.NewTicker(3 * time.Second)
		go func() {
			for {
				select {
				case <-game.cancelSendingPings:
					return
				case t := <-game.sendPingsTicker.C:
					fmt.Println("send pings to players", t)
					for _, player := range game.Players {
						if player.conn == nil {
							continue
						}
						if err := player.sendMessage(Response{
							Code:    http.StatusOK,
							Message: "ping",
						}); err != nil {
							// log.Printf("Error sending ping to client: %s\n", err.Error())
							// game.terminateGame()
							// return
							panic(fmt.Sprintf("error sending ping: %s", err.Error()))
						}
						fmt.Printf("sent ping to player: %s\n", player.Name)
					}
				}
			}
		}()
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

	return nil
}

func (g *Game) processMsg(playerId string, message []byte) error {
	var msg Message
	var err error
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error parsing message: %s\n", err.Error())
		return fmt.Errorf("error parsing message")
	}
	switch g.State {
	case RecvPick:
		err = g.processPick(playerId, msg.TopicIdx, msg.ValIdx)
	case RecvBuzz:
		err = g.processBuzz(playerId, msg.IsPass)
	case RecvAns:
		err = g.processAnswer(playerId, msg.Answer)
	case RecvAnsConfirmation:
		err = g.processAnsConfirmation(playerId, msg.Confirm)
	case RecvWager:
		err = g.processWager(playerId, msg.Wager)
	case PostGame:
		err = g.processProtest(playerId, msg.ProtestFor)
	case PreGame:
		err = fmt.Errorf("received unexpected message")
	}
	return err
}

func (g *Game) processPick(playerId string, topicIdx, valIdx int) error {
	g.cancelPickTimeout()
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

func (g *Game) processBuzz(playerId string, isPass bool) error {
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
				Code:      http.StatusOK,
				Message:   "You passed",
				Game:      g,
				CurPlayer: player,
			})
		}
		g.cancelBuzzTimeout()
		return g.skipQuestion()
	}
	g.cancelBuzzTimeout()
	g.setState(RecvAns, player.Id)
	return g.messageAllPlayers("Player buzzed")
}

func (g *Game) processAnswer(playerId, answer string) error {
	cancelAnswerTimeout := g.cancelAnswerTimeout[playerId]
	cancelAnswerTimeout()
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if !player.CanAnswer {
		return fmt.Errorf("player cannot answer")
	}
	isCorrect := g.CurQuestion.checkAnswer(answer)
	if g.Round == FinalRound {
		return g.processFinalRoundAns(player, isCorrect, answer)
	}
	g.AnsCorrectness = isCorrect
	g.LastAnswer = answer
	g.LastAnswerer = player
	g.setState(RecvAnsConfirmation, "")
	return g.messageAllPlayers("Player answered")
}

func (g *Game) processAnswerTimeout(playerId string) error {
	cancelAnswerTimeout := g.cancelAnswerTimeout[playerId]
	cancelAnswerTimeout()
	player := g.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if !player.CanAnswer {
		return fmt.Errorf("player cannot answer")
	}
	isCorrect := false
	if g.Round == FinalRound {
		return g.processFinalRoundAns(player, isCorrect, "answer-timeout")
	}
	return g.nextQuestion(player, isCorrect)
}

func (g *Game) processFinalRoundAns(player *Player, isCorrect bool, answer string) error {
	player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
	g.FinalAnswersRecvd++
	player.CanAnswer = false
	player.FinalAnswer = answer
	player.FinalCorrect = isCorrect
	if g.roundEnded() {
		g.setState(PostGame, "")
		return g.messageAllPlayers("Final round ended")
	}
	return player.sendMessage(Response{
		Code:      http.StatusOK,
		Message:   "You answered",
		Game:      g,
		CurPlayer: player,
	})
}

func (g *Game) processAnsConfirmation(playerId string, confirm bool) error {
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
			Code:      http.StatusOK,
			Message:   "You confirmed",
			Game:      g,
			CurPlayer: player,
		})
	}
	g.cancelConfirmationTimeout()
	isCorrect := (g.AnsCorrectness && g.Confirmations == 2) || (!g.AnsCorrectness && g.Challenges == 2)
	return g.nextQuestion(g.LastAnswerer, isCorrect)
}

func (g *Game) processWager(playerId string, wager int) error {
	cancelWagerTimeout := g.cancelWagerTimeout[playerId]
	cancelWagerTimeout()
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
				Code:      http.StatusOK,
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

func (g *Game) processProtest(playerId, protestFor string) error {
	protestForPlayer := g.getPlayerById(protestFor)
	protestByPlayer := g.getPlayerById(playerId)
	if protestForPlayer == nil || protestByPlayer == nil {
		return fmt.Errorf("player not found")
	}
	if _, ok := protestForPlayer.FinalProtestors[protestByPlayer.Id]; ok {
		return nil
	}
	protestForPlayer.FinalProtestors[protestByPlayer.Id] = true
	if len(protestForPlayer.FinalProtestors) != 2 {
		return protestByPlayer.sendMessage(Response{
			Code:      http.StatusOK,
			Message:   "You protested for " + protestForPlayer.Name,
			Game:      g,
			CurPlayer: protestByPlayer,
		})
	}
	if protestForPlayer.FinalCorrect {
		protestForPlayer.Score -= 2 * protestForPlayer.FinalWager
	} else {
		protestForPlayer.Score += 2 * protestForPlayer.FinalWager
	}
	g.setState(PostGame, "")
	return g.messageAllPlayers("Final Jeopardy result changed")
}

func (g *Game) startTimeout(ctx context.Context, timeout time.Duration, playerId string, processTimeout func(playerId string) error) {
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
	defer timeoutCancel()
	select {
	case <-ctx.Done():
		return
	case <-timeoutCtx.Done():
		if err := processTimeout(playerId); err != nil {
			log.Printf("Unexpected error after timeout: %s\n", err)
			g.terminateGame()
		}
		return
	}
}

func (g *Game) setState(state GameState, id string) {
	switch state {
	case RecvPick:
		for _, player := range g.Players {
			player.updateActions(player.Id == id, false, false, false, false)
		}
		ctx, cancel := context.WithCancel(context.Background())
		g.cancelPickTimeout = cancel
		go g.startTimeout(ctx, pickQuestionTimeout, "", func(_ string) error {
			topicIdx, valIdx := g.firstAvailableQuestion()
			return g.processPick(id, topicIdx, valIdx)
		})
	case RecvBuzz:
		for _, player := range g.Players {
			player.updateActions(false, player.canBuzz(g.GuessedWrong), false, false, false)
		}
		ctx, cancel := context.WithCancel(context.Background())
		g.cancelBuzzTimeout = cancel
		go g.startTimeout(ctx, buzzInTimeout, "", func(_ string) error { return g.skipQuestion() })
	case RecvAns:
		for _, player := range g.Players {
			player.updateActions(false, false, player.Id == id, false, false)
			if g.Round == FinalRound {
				player.CanAnswer = player.Score > 0
			}
		}
		for _, player := range g.Players {
			if !player.CanAnswer {
				continue
			}
			ctx, cancel := context.WithCancel(context.Background())
			g.cancelAnswerTimeout[player.Id] = cancel
			answerTimeout := defaultAnsTimeout
			if g.CurQuestion.DailyDouble {
				answerTimeout = dailyDoubleAnsTimeout
			} else if g.Round == FinalRound {
				answerTimeout = finalJeopardyAnsTimeout
			}
			go g.startTimeout(ctx, answerTimeout, player.Id, g.processAnswerTimeout)
		}
	case RecvAnsConfirmation:
		for _, player := range g.Players {
			player.updateActions(false, false, false, false, true)
		}
		ctx, cancel := context.WithCancel(context.Background())
		g.cancelConfirmationTimeout = cancel
		go g.startTimeout(ctx, confirmAnsTimeout, "", func(_ string) error {
			return g.nextQuestion(g.LastAnswerer, g.AnsCorrectness)
		})
	case RecvWager:
		for _, player := range g.Players {
			player.updateActions(false, false, false, player.Id == id, false)
			if g.Round == FinalRound {
				player.CanWager = player.Score > 0
			}
		}
		for _, player := range g.Players {
			if !player.CanWager {
				continue
			}
			ctx, cancel := context.WithCancel(context.Background())
			g.cancelWagerTimeout[player.Id] = cancel
			wagerTimeout := dailyDoubleWagerTimeout
			if g.Round == FinalRound {
				wagerTimeout = finalJeopardyWagerTimeout
			}
			go g.startTimeout(ctx, wagerTimeout, player.Id, func(playerId string) error {
				wager := 5
				if g.Round == FinalRound {
					wager = 0
				}
				return g.processWager(playerId, wager)
			})
		}
	case PreGame, PostGame:
		for _, player := range g.Players {
			player.updateActions(false, false, false, false, false)
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
	g.cancelSendingPings <- true
	g.sendPingsTicker.Stop()
	if err := g.setQuestions(); err != nil {
		return err
	}
	g.setState(RecvPick, g.Players[0].Id)
	// for i := range g.Players {
	// 	// random score between 1000 and 5000
	// 	g.Players[i].Score = (rand.Intn(5) + 1) * 1000
	// }
	// g.startFinalRound()

	for _, player := range g.Players {
		go func(player *Player) {
			// TODO: USE A CHANNEL TO WAIT ON A MESSAGE OR TO END THE GAME
			for {
				msg, err := player.readMessage()
				if err != nil {
					player.closeConnWithMsg("Error reading message from WebSocket")
					return
				}
				err = g.processMsg(player.Id, msg)
				if err != nil {
					log.Printf("Error handling request: %s\n", err.Error())
					player.closeConnWithMsg(err.Error())
					return
				}
			}
		}(player)
	}

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
	if g.cancelPickTimeout != nil {
		g.cancelPickTimeout()
	}
	if g.cancelBuzzTimeout != nil {
		g.cancelBuzzTimeout()
	}
	if g.cancelConfirmationTimeout != nil {
		g.cancelConfirmationTimeout()
	}
	for _, player := range g.Players {
		cancelAnswerTimeout, ok := g.cancelAnswerTimeout[player.Id]
		if ok {
			cancelAnswerTimeout()
		}
		cancelWagerTimeout, ok := g.cancelWagerTimeout[player.Id]
		if ok {
			cancelWagerTimeout()
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
			Code:      http.StatusOK,
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
