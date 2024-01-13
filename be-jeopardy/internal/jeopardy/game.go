package jeopardy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

type (
	Game struct {
		Name           string     `json:"name"`
		State          GameState  `json:"state"`
		Round          RoundState `json:"round"`
		FirstRound     []Topic    `json:"firstRound"`
		SecondRound    []Topic    `json:"secondRound"`
		FinalQuestion  Question   `json:"finalQuestion"`
		CurQuestion    Question   `json:"curQuestion"`
		Players        []*Player  `json:"players"`
		LastToPick     *Player    `json:"lastToPick"`
		LastToBuzz     *Player    `json:"lastToBuzz"`
		LastToAnswer   *Player    `json:"lastToAnswer"`
		LastAnswer     string     `json:"lastAnswer"`
		AnsCorrectness bool       `json:"ansCorrectness"`
		GuessedWrong   []string   `json:"guessedWrong"`
		Passed         []string   `json:"passed"`
		Confirmers     []string   `json:"confirmations"`
		Challengers    []string   `json:"challenges"`
		NumFinalWagers int        `json:"numFinalWagers"`
		FinalWagers    []string   `json:"finalWagers"`
		FinalAnswers   []string   `json:"finalAnswers"`
		Paused         bool       `json:"paused"`

		cancelPickTimeout context.CancelFunc
		cancelBuzzTimeout context.CancelFunc
		cancelVoteTimeout context.CancelFunc

		msgChan  chan Message
		stopChan chan *Player
	}

	Message struct {
		Player *Player
		PickMessage
		BuzzMessage
		AnswerMessage
		VoteMessage
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

	VoteMessage struct {
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
	RecvVote
	PostGame
)

type RoundState int

const (
	FirstRound RoundState = iota
	SecondRound
	FinalRound
)

const numPlayers = 3

func NewGame(name string) (*Game, error) {
	game := &Game{
		State:             PreGame,
		Players:           []*Player{},
		Round:             FirstRound,
		Name:              name,
		cancelPickTimeout: func() {},
		cancelBuzzTimeout: func() {},
		cancelVoteTimeout: func() {},
		msgChan:           make(chan Message),
		stopChan:          make(chan *Player),
	}
	if err := game.setQuestions(); err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case msg := <-game.msgChan:
				if err := game.processMsg(msg); err != nil {
					log.Errorf("Error processing message: %s\n", err.Error())
				}
			case player := <-game.stopChan:
				log.Infof("Stopping game %s\n", game.Name)
				game.stopGame(player)
			}
		}
	}()
	return game, nil
}

func getPlayerGame(playerId string) *Game {
	return playerGames[playerId]
}

func (g *Game) stopGame(player *Player) {
	g.Paused = true
	g.cancelPickTimeout()
	g.cancelBuzzTimeout()
	g.cancelVoteTimeout()
	player.stopSendingPings <- true
	player.Conn = nil
	for _, p := range g.Players {
		p.stopPlayer()
	}
	g.messageAllPlayers(fmt.Sprintf("Player %s left the game", player.Name))
}

func (g *Game) processMsg(msg Message) error {
	player := msg.Player
	if g.Paused {
		log.Infof("Ignoring message from player %s because game is paused\n", player.Name)
		return nil
	}
	var err error
	switch g.State {
	case RecvPick:
		log.Infof("Player %s picked\n", player.Name)
		err = g.processPick(player, msg.TopicIdx, msg.ValIdx)
	case RecvBuzz:
		log.Infof("Player %s buzzed\n", player.Name)
		err = g.processBuzz(player, msg.IsPass)
	case RecvAns:
		log.Infof("Player %s answered\n", player.Name)
		err = g.processAnswer(player, msg.Answer)
	case RecvVote:
		log.Infof("Player %s voted\n", player.Name)
		err = g.processVote(player, msg.Confirm)
	case RecvWager:
		log.Infof("Player %s wagered\n", player.Name)
		err = g.processWager(player, msg.Wager)
	case PostGame:
		log.Infof("Player %s protested\n", player.Name)
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
	g.LastToPick = player
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
		g.Passed = append(g.Passed, player.Id)
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
	g.LastToBuzz = player
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
	g.LastToAnswer = player
	g.setState(RecvVote, &Player{})
	return g.messageAllPlayers("Player answered")
}

func (g *Game) processVote(player *Player, confirm bool) error {
	if !player.CanVote {
		return fmt.Errorf("player cannot vote")
	}
	player.CanVote = false
	if confirm {
		g.Confirmers = append(g.Confirmers, player.Id)
	} else {
		g.Challengers = append(g.Challengers, player.Id)
	}
	if len(g.Confirmers) != 2 && len(g.Challengers) != 2 {
		return player.sendMessage(Response{
			Code:      http.StatusOK,
			Message:   "You voted",
			Game:      g,
			CurPlayer: player,
		})
	}
	g.cancelVoteTimeout()
	isCorrect := (g.AnsCorrectness && len(g.Confirmers) == 2) || (!g.AnsCorrectness && len(g.Challengers) == 2)
	return g.nextQuestion(g.LastToAnswer, isCorrect)
}

func (g *Game) processWager(player *Player, wager int) error {
	player.cancelWagerTimeout()
	if !player.CanWager {
		return fmt.Errorf("player cannot wager")
	}
	if min, max, ok := g.validWager(wager, player.Score); !ok {
		g.startWagerTimeout(player)
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
		g.FinalWagers = append(g.FinalWagers, player.Id)
		if len(g.FinalWagers) != g.NumFinalWagers {
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

func (g *Game) processFinalRoundAns(player *Player, isCorrect bool, answer string) error {
	player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
	g.FinalAnswers = append(g.FinalAnswers, player.Id)
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

func (g *Game) setState(state GameState, player *Player) {
	switch state {
	case RecvPick:
		for _, p := range g.Players {
			p.updateActions(p.Id == player.Id, false, false, false, false)
		}
		g.startPickTimeout(player)
	case RecvBuzz:
		for _, p := range g.Players {
			p.updateActions(false, p.canBuzz(g.GuessedWrong, g.Passed), false, false, false)
		}
		g.startBuzzTimeout(player)
	case RecvAns:
		for _, p := range g.Players {
			canAnswer := p.Id == player.Id
			if g.Round == FinalRound {
				canAnswer = p.Score > 0 && !p.inLists(g.FinalAnswers)
			}
			p.updateActions(false, false, canAnswer, false, false)
		}
		for _, p := range g.Players {
			if !p.CanAnswer {
				continue
			}
			g.startAnswerTimeout(p)
		}
	case RecvVote:
		for _, p := range g.Players {
			p.updateActions(false, false, false, false, p.canVote(g.Confirmers, g.Challengers))
		}
		g.startVoteTimeout(player)
	case RecvWager:
		for _, p := range g.Players {
			canWager := p.Id == player.Id
			if g.Round == FinalRound {
				canWager = p.Score > 0 && !p.inLists(g.FinalWagers)
			}
			p.updateActions(false, false, false, canWager, false)
		}
		for _, p := range g.Players {
			if !p.CanWager {
				continue
			}
			g.startWagerTimeout(p)
		}
	case PreGame, PostGame:
		for _, p := range g.Players {
			p.updateActions(false, false, false, false, false)
		}
	}
	g.State = state
}

func (g *Game) startGame() {
	g.Paused = false
	state, player := g.State, &Player{}
	if state == PreGame {
		state, player = RecvPick, g.Players[0]
		// state, player = PostGame, &Player{}
	} else if state == RecvWager && g.Round != FinalRound {
		player = g.LastToPick
	} else if state == RecvPick {
		for _, p := range g.Players {
			if p.CanPick {
				player = p
			}
		}
	} else if state == RecvAns && g.Round != FinalRound {
		for _, p := range g.Players {
			if p.CanAnswer {
				player = p
			}
		}
	}
	g.setState(state, player)
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
		g.setState(RecvPick, g.LastToPick)
		msg = "All players guessed wrong"
	} else if isCorrect || g.CurQuestion.DailyDouble {
		g.resetGuesses()
		g.setState(RecvPick, player)
		msg = "Question is complete"
	} else {
		g.Confirmers = []string{}
		g.Challengers = []string{}
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
		g.setState(RecvPick, g.LastToPick)
		msg = "Question unanswered"
	}
	return g.messageAllPlayers(msg)
}

func (g *Game) resetGuesses() {
	g.GuessedWrong = []string{}
	g.Passed = []string{}
	g.Confirmers = []string{}
	g.Challengers = []string{}
}

func (g *Game) messageAllPlayers(msg string) error {
	for _, player := range g.Players {
		if err := player.sendMessage(Response{
			Code:      http.StatusOK,
			Message:   msg,
			Game:      g,
			CurPlayer: player,
		}); err != nil {
			log.Errorf("Error sending message to player %s while messaging all players: %s, stopping game", player.Name, err.Error())
			g.stopGame(player)
		}
	}
	return nil
}

func (g *Game) getPlayerById(id string) *Player {
	for _, player := range g.Players {
		if player.Id == id {
			return player
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

func (g *Game) noPlayerCanBuzz() bool {
	return len(g.Passed)+len(g.GuessedWrong) == numPlayers
}

func (g *Game) roundEnded() bool {
	if g.Round == FinalRound {
		return len(g.FinalAnswers) == g.NumFinalWagers
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
