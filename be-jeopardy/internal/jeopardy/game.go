package jeopardy

import (
	"context"
	"fmt"
	"time"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/socket"
)

type (
	Game struct {
		Name           string       `json:"name"`
		State          GameState    `json:"state"`
		Round          RoundState   `json:"round"`
		FirstRound     []Category   `json:"firstRound"`
		SecondRound    []Category   `json:"secondRound"`
		FinalQuestion  Question     `json:"finalQuestion"`
		CurQuestion    Question     `json:"curQuestion"`
		Players        []GamePlayer `json:"players"`
		LastToPick     GamePlayer   `json:"lastToPick"`
		AnsCorrectness bool         `json:"ansCorrectness"`
		GuessedWrong   []string     `json:"guessedWrong"`
		Passed         []string     `json:"passed"`
		Confirmers     []string     `json:"confirmations"`
		Challengers    []string     `json:"challenges"`
		DisputePicker  GamePlayer   `json:"disputePicker"`
		Disputers      int          `json:"disputes"`
		NonDisputers   int          `json:"nonDisputes"`
		NumFinalWagers int          `json:"numFinalWagers"`
		FinalWagers    []string     `json:"finalWagers"`
		FinalAnswers   []string     `json:"finalAnswers"`
		Paused         bool         `json:"paused"`
		PausedState    GameState    `json:"pausedState"`
		PausedAt       time.Time    `json:"pausedAt"`

		StartBuzzCountdown        bool `json:"startBuzzCountdown"`
		StartFinalAnswerCountdown bool `json:"startFinalAnswerCountdown"`
		StartFinalWagerCountdown  bool `json:"startFinalWagerCountdown"`

		cancelBoardIntroTimeout context.CancelFunc
		cancelPickTimeout       context.CancelFunc
		cancelBuzzTimeout       context.CancelFunc
		cancelVoteTimeout       context.CancelFunc
		cancelDisputeTimeout    context.CancelFunc

		msgChan        chan Message
		disconnectChan chan GamePlayer
		restartChan    chan bool
		chatChan       chan ChatMessage

		questionDB QuestionDB
	}

	QuestionDB interface {
		GetQuestions() ([]db.Question, error)
		AddAlternative(alternative, answer string) error
		Close() error
	}

	Message struct {
		Player GamePlayer
		State  GameState `json:"state"`

		CatIdx     int    `json:"catIdx"`
		ValIdx     int    `json:"valIdx"`
		IsPass     bool   `json:"isPass"`
		Wager      int    `json:"wager"`
		Answer     string `json:"answer"`
		Confirm    bool   `json:"confirm"`
		Dispute    bool   `json:"dispute"`
		ProtestFor string `json:"protestFor"`

		Pause int `json:"pause"` // 1 is pause, -1 is resume
	}

	Response struct {
		Code      int        `json:"code"`
		Token     string     `json:"token,omitempty"`
		Message   string     `json:"message"`
		Game      *Game      `json:"game,omitempty"`
		CurPlayer GamePlayer `json:"curPlayer,omitempty"`
	}
)

type GameState int

const (
	PreGame GameState = iota
	BoardIntro
	RecvPick
	RecvBuzz
	RecvWager
	RecvAns
	RecvVote
	RecvDispute
	PostGame
)

type RoundState int

const (
	FirstRound RoundState = iota
	SecondRound
	FinalRound
)

const numPlayers = 3

func NewGame(db QuestionDB) (*Game, error) {
	game := &Game{
		State:                   PreGame,
		Players:                 []GamePlayer{},
		Round:                   FirstRound,
		Name:                    genGameCode(),
		LastToPick:              &Player{},
		cancelBoardIntroTimeout: func() {},
		cancelPickTimeout:       func() {},
		cancelBuzzTimeout:       func() {},
		cancelVoteTimeout:       func() {},
		msgChan:                 make(chan Message),
		disconnectChan:          make(chan GamePlayer),
		restartChan:             make(chan bool),
		chatChan:                make(chan ChatMessage),
		questionDB:              db,
	}
	if err := game.setQuestions(); err != nil {
		return nil, err
	}
	game.processMessages()
	game.processChatMessages()
	return game, nil
}

func (g *Game) processMessages() {
	go func() {
		for {
			select {
			case msg := <-g.msgChan:
				if err := g.processMsg(msg); err != nil {
					log.Errorf("Error processing message: %s\n", err.Error())
				}
			case player := <-g.disconnectChan:
				log.Infof("Stopping game %s\n", g.Name)
				g.disconnectPlayer(player)
			case <-g.restartChan:
				log.Infof("Restarting game %s\n", g.Name)
				g.restartGame()
			}
		}
	}()
}

func (g *Game) restartGame() {
	g.State = PreGame
	g.Round = FirstRound
	g.LastToPick = &Player{}
	g.CurQuestion = Question{}
	g.AnsCorrectness = false
	g.GuessedWrong = []string{}
	g.Passed = []string{}
	g.Confirmers = []string{}
	g.Challengers = []string{}
	g.Disputers = 0
	g.NonDisputers = 0
	g.NumFinalWagers = 0
	g.FinalWagers = []string{}
	g.FinalAnswers = []string{}
	g.setQuestions()
	for _, p := range g.Players {
		p.resetPlayer()
	}
	g.setState(BoardIntro, g.Players[0])
	g.messageAllPlayers("We are ready to play")
}

func (g *Game) pauseGame() {
	g.PausedAt = time.Now()
	g.Paused = true
	g.PausedState = g.State
	g.cancelBoardIntroTimeout()
	g.cancelPickTimeout()
	g.cancelBuzzTimeout()
	g.cancelVoteTimeout()
	for _, p := range g.Players {
		p.pausePlayer()
	}
}

func (g *Game) disconnectPlayer(player GamePlayer) {
	g.pauseGame()
	if g.State != PostGame {
		g.State = PreGame
	}
	player.endConnections()
	player.setPlayAgain(false)
	g.messageAllPlayers("Player %s disconnected from the game", player.name())
	endGame := true
	for _, p := range g.Players {
		if p.conn() != nil {
			endGame = false
		}
	}
	if endGame {
		log.Infof("All players disconnected, removing game %s\n", g.Name)
		removeGame(g)
	}

}

func (g *Game) processMsg(msg Message) error {
	player := msg.Player
	if g.State != msg.State {
		log.Infof("Ignoring message from player %s because it is not for the current game state", player.name())
		return nil
	}
	if g.Paused {
		if msg.Pause == -1 {
			log.Infof("Player %s resumed the game", player.name())
			g.startGame()
			g.messageAllPlayers("Player %s resumed the game", player.name())
			return nil
		}
		log.Infof("Ignoring message from player %s because game is paused", player.name())
		return nil
	}
	if msg.Pause == 1 {
		log.Infof("Player %s paused the game", player.name())
		g.pauseGame()
		g.messageAllPlayers("Player %s paused the game", player.name())
		return nil
	}
	var err error
	switch g.State {
	case RecvPick:
		if msg.Dispute {
			log.Infof("Player %s disputed the previous question", player.name())
			err = g.initDispute(player)
		} else {
			log.Infof("Player %s picked", player.name())
			err = g.processPick(player, msg.CatIdx, msg.ValIdx)
		}
	case RecvBuzz:
		action := "buzzed"
		if msg.IsPass {
			action = "passed"
		}
		log.Infof("Player %s %s", player.name(), action)
		err = g.processBuzz(player, msg.IsPass)
	case RecvWager:
		log.Infof("Player %s wagered", player.name())
		err = g.processWager(player, msg.Wager)
	case RecvAns:
		log.Infof("Player %s answered", player.name())
		err = g.processAnswer(player, msg.Answer)
	case RecvVote:
		action := "accepted"
		if !msg.Confirm {
			action = "challenged"
		}
		log.Infof("Player %s %s", player.name(), action)
		err = g.processVote(player, msg.Confirm)
	case RecvDispute:
		action := "confirmed"
		if !msg.Dispute {
			action = "disputed"
		}
		log.Infof("Player %s %s the question", player.name(), action)
		err = g.processDispute(player, msg.Dispute)
	case PostGame:
		log.Infof("Player %s protested", player.name())
		err = g.processProtest(player, msg.ProtestFor)
	case PreGame:
		err = fmt.Errorf("received unexpected message")
	}
	return err
}

func (g *Game) getIncorrectAns(player GamePlayer) (*Answer, bool) {
	for _, ans := range g.CurQuestion.Incorrect {
		if ans.Player.id() == player.id() && !ans.HasDisputed {
			return ans, true
		}
	}
	return &Answer{}, false
}

func (g *Game) initDispute(player GamePlayer) error {
	ans, canDispute := g.getIncorrectAns(player)
	if !canDispute {
		return fmt.Errorf("player cannot initiate dispute")
	}
	g.cancelPickTimeout()
	for _, p := range g.Players {
		if p.canPick() {
			g.DisputePicker = p
		}
	}
	g.Disputers = 1
	ans.HasDisputed = true
	g.CurQuestion.CurDisputed = ans
	g.setState(RecvDispute, player)
	g.messageAllPlayers("Player %s disputed the answer", player.name())
	return nil
}

func (g *Game) processPick(player GamePlayer, catIdx, valIdx int) error {
	if !player.canPick() {
		return fmt.Errorf("player cannot pick")
	}
	if catIdx < 0 || valIdx < 0 || catIdx >= numCategories || valIdx >= numQuestions {
		return fmt.Errorf("invalid question pick")
	}
	g.cancelPickTimeout()
	curRound := g.FirstRound
	if g.Round == SecondRound {
		curRound = g.SecondRound
	}
	curQuestion := curRound[catIdx].Questions[valIdx]
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
	g.messageAllPlayers(msg)
	return nil
}

func (g *Game) processBuzz(player GamePlayer, isPass bool) error {
	if !player.canBuzz() {
		return fmt.Errorf("player cannot buzz")
	}
	if isPass {
		g.Passed = append(g.Passed, player.id())
		player.setCanBuzz(false)
		if g.noPlayerCanBuzz() {
			g.cancelBuzzTimeout()
			g.skipQuestion()
			return nil
		}
		g.StartBuzzCountdown = false
		_ = player.sendMessage(Response{
			Code:      socket.Ok,
			Message:   "You passed",
			Game:      g,
			CurPlayer: player,
		})
		return nil
	}
	g.cancelBuzzTimeout()
	g.setState(RecvAns, player)
	g.messageAllPlayers("Player buzzed")
	return nil
}

func (g *Game) processAnswer(player GamePlayer, answer string) error {
	if !player.canAnswer() {
		return fmt.Errorf("player cannot answer")
	}
	player.cancelAnswerTimeout()
	isCorrect := g.CurQuestion.checkAnswer(answer)
	if g.Round == FinalRound {
		return g.processFinalRoundAns(player, isCorrect, answer)
	}
	g.AnsCorrectness = isCorrect
	g.CurQuestion.CurAns = &Answer{
		Player: player,
		Answer: answer,
	}
	g.setState(RecvVote, &Player{})
	g.messageAllPlayers("Player answered")
	return nil
}

func (g *Game) processVote(player GamePlayer, confirm bool) error {
	if !player.canVote() {
		return fmt.Errorf("player cannot vote")
	}
	player.setCanVote(false)
	if confirm {
		g.Confirmers = append(g.Confirmers, player.id())
	} else {
		g.Challengers = append(g.Challengers, player.id())
	}
	if len(g.Confirmers) != 2 && len(g.Challengers) != 2 {
		_ = player.sendMessage(Response{
			Code:      socket.Ok,
			Message:   "You voted",
			Game:      g,
			CurPlayer: player,
		})
		return nil
	}
	g.cancelVoteTimeout()
	isCorrect := (g.AnsCorrectness && len(g.Confirmers) == 2) || (!g.AnsCorrectness && len(g.Challengers) == 2)
	if !g.AnsCorrectness && len(g.Challengers) == 2 {
		if err := g.questionDB.AddAlternative(g.CurQuestion.CurAns.Answer, g.CurQuestion.Answer); err != nil {
			log.Errorf("Error adding alternative: %s", err.Error())
		}
	}
	if isCorrect {
		g.CurQuestion.Correct = g.CurQuestion.CurAns
	} else {
		g.CurQuestion.Incorrect = append(g.CurQuestion.Incorrect, g.CurQuestion.CurAns)
	}
	g.nextQuestion(g.CurQuestion.CurAns.Player, isCorrect)
	return nil
}

func (g *Game) processDispute(player GamePlayer, dispute bool) error {
	if !player.canDispute() {
		return fmt.Errorf("player cannot dispute")
	}
	player.setCanDispute(false)
	if dispute {
		g.Disputers++
	} else {
		g.NonDisputers++

	}
	if g.Disputers != 2 && g.NonDisputers != 2 {
		_ = player.sendMessage(Response{
			Code:      socket.Ok,
			Message:   "You disputed",
			Game:      g,
			CurPlayer: player,
		})
		return nil
	}
	g.cancelDisputeTimeout()
	nextPicker := g.DisputePicker
	if g.Disputers > g.NonDisputers {
		g.CurQuestion.CurDisputed.Player.addToScore(2 * g.CurQuestion.Value)
		if err := g.questionDB.AddAlternative(g.CurQuestion.CurDisputed.Answer, g.CurQuestion.Answer); err != nil {
			log.Errorf("Error adding alternative: %s", err.Error())
		}
		nextPicker = g.CurQuestion.CurDisputed.Player
	}
	g.Disputers = 0
	g.NonDisputers = 0
	g.setState(RecvPick, nextPicker)
	g.messageAllPlayers("Dispute resolved")
	return nil
}

func (g *Game) processWager(player GamePlayer, wager int) error {
	if !player.canWager() {
		return fmt.Errorf("player cannot wager")
	}
	player.cancelWagerTimeout()
	if min, max, ok := g.validWager(wager, player.score()); !ok {
		g.startWagerTimeout(player)
		_ = player.sendMessage(Response{
			Code:      socket.BadRequest,
			Message:   fmt.Sprintf("invalid wager, must be between %d and %d", min, max),
			Game:      g,
			CurPlayer: player,
		})
		return nil
	}
	var msg string
	if g.Round == FinalRound {
		player.setFinalWager(wager)
		player.setCanWager(false)
		g.FinalWagers = append(g.FinalWagers, player.id())
		if len(g.FinalWagers) != g.NumFinalWagers {
			g.StartFinalWagerCountdown = false
			_ = player.sendMessage(Response{
				Code:      socket.Ok,
				Message:   "You wagered",
				Game:      g,
				CurPlayer: player,
			})
			return nil
		}
		g.setState(RecvAns, &Player{})
		msg = "All wagers received"
	} else {
		// daily double
		g.CurQuestion.Value = wager
		g.setState(RecvAns, player)
		msg = "Player wagered"
	}
	g.messageAllPlayers(msg)
	return nil
}

func (g *Game) processProtest(protestByPlayer GamePlayer, protestFor string) error {
	protestForPlayer, err := g.getPlayerById(protestFor)
	if err != nil {
		return err
	}
	if _, ok := protestForPlayer.finalProtestors()[protestByPlayer.id()]; ok {
		return nil
	}
	protestForPlayer.addFinalProtestor(protestByPlayer.id())
	if len(protestForPlayer.finalProtestors()) != numPlayers/2+1 {
		_ = protestByPlayer.sendMessage(Response{
			Code:      socket.Ok,
			Message:   "You protested for " + protestForPlayer.name(),
			Game:      g,
			CurPlayer: protestByPlayer,
		})
		return nil
	}
	if protestForPlayer.finalCorrect() {
		protestForPlayer.addToScore(-2 * protestForPlayer.finalWager())
	} else {
		protestForPlayer.addToScore(2 * protestForPlayer.finalWager())
	}
	protestForPlayer.setFinalCorrect(!protestForPlayer.finalCorrect())
	g.setState(PostGame, &Player{})
	g.messageAllPlayers("Final Jeopardy result changed")
	return nil
}

func (g *Game) processFinalRoundAns(player GamePlayer, isCorrect bool, answer string) error {
	player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
	g.FinalAnswers = append(g.FinalAnswers, player.id())
	player.setCanAnswer(false)
	player.setFinalAnswer(answer)
	player.setFinalCorrect(isCorrect)
	if g.roundEnded() {
		g.setState(PostGame, &Player{})
		g.messageAllPlayers("Final round ended")
		return nil
	}
	g.StartFinalAnswerCountdown = false
	_ = player.sendMessage(Response{
		Code:      socket.Ok,
		Message:   "You answered",
		Game:      g,
		CurPlayer: player,
	})
	return nil
}

func (g *Game) setState(state GameState, player GamePlayer) {
	switch state {
	case BoardIntro:
		for _, p := range g.Players {
			p.updateActions(p.id() == player.id(), false, false, false, false)
		}
		g.startBoardIntroTimeout()
	case RecvPick:
		for _, p := range g.Players {
			p.updateActions(p.id() == player.id(), false, false, false, false)
		}
		g.startPickTimeout(player)
	case RecvBuzz:
		for _, p := range g.Players {
			p.updateActions(false, !inLists(p.id(), g.GuessedWrong, g.Passed), false, false, false)
		}
		g.startBuzzTimeout()
	case RecvAns:
		for _, p := range g.Players {
			canAnswer := p.id() == player.id()
			if g.Round == FinalRound {
				canAnswer = p.score() > 0 && !inLists(p.id(), g.FinalAnswers)
			}
			p.updateActions(false, false, canAnswer, false, false)
		}
		for _, p := range g.Players {
			if !p.canAnswer() {
				continue
			}
			g.startAnswerTimeout(p)
		}
	case RecvVote:
		for _, p := range g.Players {
			p.updateActions(false, false, false, false, !inLists(p.id(), g.Confirmers, g.Challengers))
		}
		g.startVoteTimeout()
	case RecvWager:
		for _, p := range g.Players {
			canWager := p.id() == player.id()
			if g.Round == FinalRound {
				canWager = p.score() > 0 && !inLists(p.id(), g.FinalWagers)
			}
			p.updateActions(false, false, false, canWager, false)
		}
		for _, p := range g.Players {
			if !p.canWager() {
				continue
			}
			g.startWagerTimeout(p)
		}
	case RecvDispute:
		for _, p := range g.Players {
			p.updateActions(false, false, false, false, false)
			p.setCanDispute(p.id() != player.id())
		}
		g.startDisputeTimeout()
	case PreGame, PostGame:
		for _, p := range g.Players {
			p.updateActions(false, false, false, false, false)
		}
	}
	g.State = state
}

func (g *Game) startGame() {
	if g.Paused {
		g.State = g.PausedState
	}
	g.Paused = false
	state := g.State
	var player GamePlayer
	if state == PreGame || state == BoardIntro {
		state, player = RecvPick, g.Players[0]
	} else if state == RecvWager && g.Round != FinalRound {
		player = g.LastToPick
	} else if state == RecvPick {
		for _, p := range g.Players {
			if p.canPick() {
				player = p
			}
		}
	} else if state == RecvAns && g.Round != FinalRound {
		for _, p := range g.Players {
			if p.canAnswer() {
				player = p
			}
		}
	}
	g.setState(state, player)
}

func (g *Game) startSecondRound() {
	g.Round = SecondRound
	g.resetGuesses()
	g.setState(BoardIntro, g.lowestPlayer())
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

func (g *Game) nextQuestion(player GamePlayer, isCorrect bool) {
	player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
	if !isCorrect {
		g.GuessedWrong = append(g.GuessedWrong, player.id())
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
		g.Disputers = 0
		g.NonDisputers = 0
		g.setState(RecvBuzz, &Player{})
		msg = "Player answered incorrectly"
	}
	g.messageAllPlayers(msg)
}

func (g *Game) skipQuestion() {
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
	g.messageAllPlayers(msg)
}

func (g *Game) resetGuesses() {
	g.GuessedWrong = []string{}
	g.Passed = []string{}
	g.Confirmers = []string{}
	g.Challengers = []string{}
	g.Disputers = 0
	g.NonDisputers = 0
}

func (g *Game) messageAllPlayers(msg string, args ...any) {
	for _, p := range g.Players {
		_ = p.sendMessage(Response{
			Code:      socket.Ok,
			Message:   fmt.Sprintf(msg, args...),
			Game:      g,
			CurPlayer: p,
		})
	}
}

func (g *Game) getPlayerById(id string) (GamePlayer, error) {
	for _, p := range g.Players {
		if p.id() == id {
			return p, nil
		}
	}
	return &Player{}, fmt.Errorf("Player not found in game %s", g.Name)
}

func (g *Game) allPlayersReady() bool {
	ready := 0
	for _, p := range g.Players {
		if p.conn() != nil {
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
	for _, category := range curRound {
		for _, question := range category.Questions {
			if question.CanChoose {
				return false
			}
		}
	}
	return true
}

func (g *Game) lowestPlayer() GamePlayer {
	lowest := g.Players[0]
	for _, p := range g.Players {
		if p.score() < lowest.score() {
			lowest = p
		}
	}
	return lowest
}

func (g *Game) numFinalWagers() int {
	numWagers := 0
	for _, p := range g.Players {
		if p.score() > 0 {
			numWagers++
		}
	}
	return numWagers
}

func (g *Game) roundMax() int {
	switch g.Round {
	case FirstRound:
		return 1000
	case SecondRound:
		return 2000
	}
	return 0
}

func (g *Game) validWager(wager, score int) (int, int, bool) {
	minWager := 5
	if g.Round == FinalRound {
		minWager = 0
	}
	return minWager, max(score, g.roundMax()), wager >= minWager && wager <= max(score, g.roundMax())
}

func inLists(playerId string, lists ...[]string) bool {
	for _, list := range lists {
		for _, id := range list {
			if id == playerId {
				return true
			}
		}
	}
	return false
}
