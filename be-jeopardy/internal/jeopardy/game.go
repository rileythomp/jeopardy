package jeopardy

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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
		LastPicker        *Player    `json:"lastPicker"`
		LastBuzzer        *Player    `json:"lastBuzzer"`
		LastAnswerer      *Player    `json:"lastAnswerer"`
		LastAnswer        string     `json:"lastAnswer"`
		GuessedWrong      []string   `json:"guessedWrong"`
		AnsCorrectness    bool       `json:"ansCorrectness"`
		NumFinalWagers    int        `json:"numFinalWagers"`
		FinalWagersRecvd  int        `json:"finalWagers"`
		FinalAnswersRecvd int        `json:"finalAnswers"`
		Passes            int        `json:"passes"`
		Confirmations     int        `json:"confirmations"`
		Challenges        int        `json:"challenges"`

		Name string `json:"name"`

		cancelPickTimeout         context.CancelFunc
		cancelBuzzTimeout         context.CancelFunc
		cancelConfirmationTimeout context.CancelFunc
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
	finalJeopardy = false
	numPlayers    = 3

	// pickQuestionTimeout       = 10 / 2 * time.Second
	// buzzInTimeout             = 10 / 2 * time.Second
	// defaultAnsTimeout         = 10 / 2 * time.Second
	// dailyDoubleAnsTimeout     = 10 / 2 * time.Second
	// finalJeopardyAnsTimeout   = 10 / 2 * time.Second
	// confirmAnsTimeout         = 10 / 2 * time.Second
	// dailyDoubleWagerTimeout   = 10 / 2 * time.Second
	// finalJeopardyWagerTimeout = 10 / 2 * time.Second

	pickQuestionTimeout       = 10 * time.Second
	buzzInTimeout             = 10 * time.Second
	defaultAnsTimeout         = 10 * time.Second
	dailyDoubleAnsTimeout     = 10 * time.Second
	finalJeopardyAnsTimeout   = 10 * time.Second
	confirmAnsTimeout         = 10 * time.Second
	dailyDoubleWagerTimeout   = 10 * time.Second
	finalJeopardyWagerTimeout = 10 * time.Second
)

var (
	games       = map[string]*Game{}
	playerGames = map[string]*Game{}
)

func NewGame(name string) (*Game, error) {
	game := &Game{
		State:                     PreGame,
		Players:                   []*Player{},
		Round:                     FirstRound,
		Name:                      name,
		cancelPickTimeout:         func() {},
		cancelBuzzTimeout:         func() {},
		cancelConfirmationTimeout: func() {},
	}
	if err := game.setQuestions(); err != nil {
		return nil, err
	}
	return game, nil
}

func GetGames() map[string]*Game {
	return games
}

func GetGame(playerId string) *Game {
	return playerGames[playerId]
}

func JoinGame(playerName string, gameName string) (*Game, string, error) {
	game, ok := games[gameName]
	if !ok {
		var err error
		game, err = NewGame(gameName)
		if err != nil {
			log.Printf("Error creating game: %s", err.Error())
			return &Game{}, "", fmt.Errorf("error creating game")
		}
		games[gameName] = game
	}

	playerId, err := game.addPlayer(playerName)
	if err != nil {
		log.Printf("Error adding player to game: %s", err.Error())
		return &Game{}, "", err
	}
	playerGames[playerId] = game

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
	player.Conn = conn

	msg := "Waiting for more players"
	if game.allPlayersReady() {
		if err := game.startGame(); err != nil {
			return err
		}
		msg = "We are ready to play"
	}

	player.sendPings()
	player.processMessages(game)

	if err := game.messageAllPlayers(msg); err != nil {
		return err
	}

	return nil
}

func (g *Game) stopGame(player *Player) {
	g.cancelPickTimeout()
	g.cancelBuzzTimeout()
	g.cancelConfirmationTimeout()
	player.stopSendingPings <- true
	player.Conn = nil
	for _, p := range g.Players {
		p.stopPlayer()
		// TODO: MAKE MESSAGEALLPLAYERS SAFE AND USE IT HERE
		if p.Conn != nil {
			p.sendMessage(Response{
				Code:      http.StatusOK,
				Message:   "Player " + player.Name + " left the game",
				CurPlayer: p,
				Game:      g,
			})
		}
	}
}

func (g *Game) processMsg(player *Player, message []byte) error {
	var msg Message
	var err error
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error parsing message: %s\n", err.Error())
		return fmt.Errorf("error parsing message")
	}
	switch g.State {
	case RecvPick:
		fmt.Printf("Player %s made a pick\n", player.Name)
		err = g.processPick(player, msg.TopicIdx, msg.ValIdx)
	case RecvBuzz:
		fmt.Printf("Player %s buzzed\n", player.Name)
		err = g.processBuzz(player, msg.IsPass)
	case RecvAns:
		fmt.Printf("Player %s answered\n", player.Name)
		err = g.processAnswer(player, msg.Answer)
	case RecvAnsConfirmation:
		fmt.Printf("Player %s confirmed\n", player.Name)
		err = g.processAnsConfirmation(player, msg.Confirm)
	case RecvWager:
		fmt.Printf("Player %s wagered\n", player.Name)
		err = g.processWager(player, msg.Wager)
	case PostGame:
		fmt.Printf("Player %s protested\n", player.Name)
		err = g.processProtest(player, msg.ProtestFor)
	case PreGame:
		err = fmt.Errorf("received unexpected message")
	}
	return err
}

func (g *Game) processPick(player *Player, topicIdx, valIdx int) error {
	g.cancelPickTimeout()
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
	g.LastPicker = player
	g.CurQuestion = curQuestion
	var msg string
	if curQuestion.DailyDouble {
		g.setState(RecvWager, player)
		msg = "Daily Double"
	} else {
		g.setState(RecvBuzz, &Player{})
		msg = "New Question"
	}
	return g.messageAllPlayers(msg)
}

func (g *Game) processBuzz(player *Player, isPass bool) error {
	if !player.CanBuzz {
		return fmt.Errorf("player cannot buzz")
	}
	if isPass {
		g.Passes++
		player.CanBuzz = false
		if g.noPlayerCanBuzz() {
			g.cancelBuzzTimeout()
			return g.skipQuestion()
		}
		return player.sendMessage(Response{
			Code:      http.StatusOK,
			Message:   "You passed",
			Game:      g,
			CurPlayer: player,
		})
	}
	g.LastBuzzer = player
	g.cancelBuzzTimeout()
	g.setState(RecvAns, player)
	return g.messageAllPlayers("Player buzzed")
}

func (g *Game) processAnswer(player *Player, answer string) error {
	player.cancelAnswerTimeout()
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
	g.setState(RecvAnsConfirmation, &Player{})
	return g.messageAllPlayers("Player answered")
}

func (g *Game) processAnswerTimeout(player *Player) error {
	player.cancelAnswerTimeout()
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
		g.setState(PostGame, &Player{})
		return g.messageAllPlayers("Final round ended")
	}
	return player.sendMessage(Response{
		Code:      http.StatusOK,
		Message:   "You answered",
		Game:      g,
		CurPlayer: player,
	})
}

func (g *Game) processAnsConfirmation(player *Player, confirm bool) error {
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

func (g *Game) processWager(player *Player, wager int) error {
	player.cancelWagerTimeout()
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
		g.setState(RecvAns, &Player{})
		msg = "All wagers received"
	} else {
		// daily double
		g.CurQuestion.Value = wager
		g.setState(RecvAns, player)
		msg = "Player wagered"
	}
	return g.messageAllPlayers(msg)
}

func (g *Game) processProtest(protestByPlayer *Player, protestFor string) error {
	protestForPlayer := g.getPlayerById(protestFor)
	if protestForPlayer == nil {
		return fmt.Errorf("player not found")
	}
	if _, ok := protestForPlayer.FinalProtestors[protestByPlayer.Id]; ok {
		return nil
	}
	protestForPlayer.FinalProtestors[protestByPlayer.Id] = true
	if len(protestForPlayer.FinalProtestors) != numPlayers/2+1 {
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
	g.setState(PostGame, &Player{})
	return g.messageAllPlayers("Final Jeopardy result changed")
}

func (g *Game) startTimeout(ctx context.Context, timeout time.Duration, player *Player, processTimeout func(player *Player) error) {
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
	defer timeoutCancel()
	select {
	case <-ctx.Done():
		return
	case <-timeoutCtx.Done():
		if err := processTimeout(player); err != nil {
			log.Printf("Unexpected error after timeout for player %s: %s\n", player.Name, err)
			panic("error processing a timeout")
		}
		return
	}
}

func (g *Game) setState(state GameState, player *Player) {
	switch state {
	case RecvPick:
		for _, p := range g.Players {
			p.updateActions(p.Id == player.Id, false, false, false, false)
		}
		ctx, cancel := context.WithCancel(context.Background())
		g.cancelPickTimeout = cancel
		go g.startTimeout(ctx, pickQuestionTimeout, &Player{}, func(_ *Player) error {
			topicIdx, valIdx := g.firstAvailableQuestion()
			return g.processPick(player, topicIdx, valIdx)
		})
	case RecvBuzz:
		for _, p := range g.Players {
			p.updateActions(false, p.canBuzz(g.GuessedWrong), false, false, false)
		}
		ctx, cancel := context.WithCancel(context.Background())
		g.cancelBuzzTimeout = cancel
		go g.startTimeout(ctx, buzzInTimeout, &Player{}, func(_ *Player) error { return g.skipQuestion() })
	case RecvAns:
		for _, p := range g.Players {
			p.updateActions(false, false, p.Id == player.Id, false, false)
			if g.Round == FinalRound {
				p.CanAnswer = p.Score > 0
			}
		}
		for _, p := range g.Players {
			if !p.CanAnswer {
				continue
			}
			ctx, cancel := context.WithCancel(context.Background())
			p.cancelAnswerTimeout = cancel
			answerTimeout := defaultAnsTimeout
			if g.CurQuestion.DailyDouble {
				answerTimeout = dailyDoubleAnsTimeout
			} else if g.Round == FinalRound {
				answerTimeout = finalJeopardyAnsTimeout
			}
			go g.startTimeout(ctx, answerTimeout, p, g.processAnswerTimeout)
		}
	case RecvAnsConfirmation:
		for _, p := range g.Players {
			p.updateActions(false, false, false, false, true)
		}
		ctx, cancel := context.WithCancel(context.Background())
		g.cancelConfirmationTimeout = cancel
		go g.startTimeout(ctx, confirmAnsTimeout, &Player{}, func(_ *Player) error {
			return g.nextQuestion(g.LastAnswerer, g.AnsCorrectness)
		})
	case RecvWager:
		for _, p := range g.Players {
			p.updateActions(false, false, false, p.Id == player.Id, false)
			if g.Round == FinalRound {
				p.CanWager = p.Score > 0
			}
		}
		for _, p := range g.Players {
			if !p.CanWager {
				continue
			}
			ctx, cancel := context.WithCancel(context.Background())
			p.cancelWagerTimeout = cancel
			wagerTimeout := dailyDoubleWagerTimeout
			if g.Round == FinalRound {
				wagerTimeout = finalJeopardyWagerTimeout
			}
			go g.startTimeout(ctx, wagerTimeout, p, func(player *Player) error {
				wager := 5
				if g.Round == FinalRound {
					wager = 0
				}
				return g.processWager(player, wager)
			})
		}
	case PreGame, PostGame:
		for _, p := range g.Players {
			p.updateActions(false, false, false, false, false)
		}
	}
	g.State = state
}

func (g *Game) addPlayer(name string) (string, error) {
	for _, player := range g.Players {
		if player.Conn == nil {
			player.Name = name
			return player.Id, nil
		}
	}
	if len(g.Players) >= numPlayers {
		return "", fmt.Errorf("game is full")
	}
	player := NewPlayer(name)
	g.Players = append(g.Players, player)
	return player.Id, nil
}

func (g *Game) startGame() error {
	if finalJeopardy {
		for _, p := range g.Players {
			p.Score = (rand.Intn(5) + 1) * 1000
		}
		g.startFinalRound()
	} else {
		switch g.State {
		case PreGame:
			g.setState(RecvPick, g.Players[0])
		case RecvPick:
			g.setState(RecvPick, g.LastPicker)
		case RecvBuzz:
			g.setState(RecvBuzz, &Player{})
		case RecvWager:
			if g.Round == FinalRound {
				g.setState(RecvAns, &Player{})
			} else {
				g.setState(RecvWager, g.LastPicker)
			}
		case RecvAns:
			g.setState(RecvAns, g.LastBuzzer)
		case RecvAnsConfirmation:
			g.setState(RecvAnsConfirmation, &Player{})
		case PostGame:
			g.setState(PostGame, &Player{})
		}
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
		g.setState(PostGame, &Player{})
	} else {
		g.setState(RecvWager, &Player{})
	}
}

func (g *Game) noPlayerCanBuzz() bool {
	return g.Passes+len(g.GuessedWrong) == numPlayers
}

func (g *Game) nextQuestion(player *Player, isCorrect bool) error {
	player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
	if !isCorrect {
		g.GuessedWrong = append(g.GuessedWrong, player.Id)
	}
	if isCorrect || g.CurQuestion.DailyDouble || g.noPlayerCanBuzz() {
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
	} else if g.noPlayerCanBuzz() {
		g.resetGuesses()
		g.setState(RecvPick, g.LastPicker)
		msg = "All players guessed wrong"
	} else if isCorrect || g.CurQuestion.DailyDouble {
		g.resetGuesses()
		g.setState(RecvPick, player)
		msg = "Question is complete"
	} else {
		g.Confirmations = 0
		g.Challenges = 0
		g.setState(RecvBuzz, &Player{})
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

func (g *Game) allPlayersReady() bool {
	ready := 0
	for _, player := range g.Players {
		if player.Conn != nil {
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

func (g *Game) lowestPlayer() *Player {
	lowest := g.Players[0]
	for _, player := range g.Players {
		if player.Score < lowest.Score {
			lowest = player
		}
	}
	return lowest
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
