package jeopardy

import (
	"context"
	"fmt"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/socket"
)

type (
	Game struct {
		Name             string     `json:"name"`
		State            GameState  `json:"state"`
		Round            RoundState `json:"round"`
		FirstRound       []Category `json:"firstRound"`
		SecondRound      []Category `json:"secondRound"`
		FinalQuestion    Question   `json:"finalQuestion"`
		CurQuestion      Question   `json:"curQuestion"`
		Players          []*Player  `json:"players"`
		LastToPick       *Player    `json:"lastToPick"`
		LastToAnswer     *Player    `json:"lastToAnswer"`
		PreviousQuestion string     `json:"previousQuestion"`
		PreviousAnswer   string     `json:"previousAnswer"`
		LastAnswer       string     `json:"lastAnswer"`
		AnsCorrectness   bool       `json:"ansCorrectness"`
		GuessedWrong     []string   `json:"guessedWrong"`
		Passed           []string   `json:"passed"`
		Confirmers       []string   `json:"confirmations"`
		Challengers      []string   `json:"challenges"`
		NumFinalWagers   int        `json:"numFinalWagers"`
		FinalWagers      []string   `json:"finalWagers"`
		FinalAnswers     []string   `json:"finalAnswers"`
		Paused           bool       `json:"paused"`
		PausedState      GameState  `json:"pausedState"`

		StartBuzzCountdown        bool `json:"startBuzzCountdown"`
		StartFinalAnswerCountdown bool `json:"startFinalAnswerCountdown"`
		StartFinalWagerCountdown  bool `json:"startFinalWagerCountdown"`

		cancelBoardIntroTimeout context.CancelFunc
		cancelPickTimeout       context.CancelFunc
		cancelBuzzTimeout       context.CancelFunc
		cancelVoteTimeout       context.CancelFunc

		msgChan     chan Message
		pauseChan   chan *Player
		restartChan chan bool
		chatChan    chan ChatMessage

		questionDB QuestionDB
	}

	QuestionDB interface {
		GetQuestions() ([]db.Question, error)
		Close() error
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
		CatIdx int `json:"catIdx"`
		ValIdx int `json:"valIdx"`
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
	BoardIntro
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

func NewGame(db QuestionDB) (*Game, error) {
	game := &Game{
		State:                   PreGame,
		Players:                 []*Player{},
		Round:                   FirstRound,
		Name:                    genGameCode(),
		cancelBoardIntroTimeout: func() {},
		cancelPickTimeout:       func() {},
		cancelBuzzTimeout:       func() {},
		cancelVoteTimeout:       func() {},
		msgChan:                 make(chan Message),
		pauseChan:               make(chan *Player),
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
			case player := <-g.pauseChan:
				log.Infof("Stopping game %s\n", g.Name)
				g.pauseGame(player)
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
	g.LastToAnswer = &Player{}
	g.PreviousQuestion = ""
	g.PreviousAnswer = ""
	g.LastAnswer = ""
	g.AnsCorrectness = false
	g.GuessedWrong = []string{}
	g.Passed = []string{}
	g.Confirmers = []string{}
	g.Challengers = []string{}
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

func (g *Game) pauseGame(player *Player) {
	g.Paused = true
	g.PausedState = g.State
	if g.State != PostGame {
		g.State = PreGame
	}
	g.cancelBoardIntroTimeout()
	g.cancelPickTimeout()
	g.cancelBuzzTimeout()
	g.cancelVoteTimeout()
	player.Conn = nil
	player.ChatConn = nil
	player.PlayAgain = false
	for _, p := range g.Players {
		p.pausePlayer()
	}
	g.messageAllPlayers(fmt.Sprintf("Player %s left the game", player.Name))
	endGame := true
	for _, p := range g.Players {
		if p.Conn != nil {
			endGame = false
		}
	}
	if endGame {
		log.Infof("All players disconnected, removing game %s\n", g.Name)
		if err := g.questionDB.Close(); err != nil {
			log.Errorf("Error closing question db: %s\n", err.Error())
		}
		delete(publicGames, g.Name)
		delete(privateGames, g.Name)
		for _, p := range g.Players {
			delete(playerGames, p.Id)
		}
	}

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
		err = g.processPick(player, msg.CatIdx, msg.ValIdx)
	case RecvBuzz:
		action := "buzzed"
		if msg.IsPass {
			action = "passed"
		}
		log.Infof("Player %s %s\n", player.Name, action)
		err = g.processBuzz(player, msg.IsPass)
	case RecvAns:
		log.Infof("Player %s answered\n", player.Name)
		err = g.processAnswer(player, msg.Answer)
	case RecvVote:
		action := "accepted"
		if !msg.Confirm {
			action = "challenged"
		}
		log.Infof("Player %s %s\n", player.Name, action)
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

func (g *Game) processPick(player *Player, catIdx, valIdx int) error {
	if !player.CanPick {
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
	g.PreviousQuestion = g.CurQuestion.Clue
	g.PreviousAnswer = g.CurQuestion.Answer
	g.messageAllPlayers(msg)
	return nil
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

func (g *Game) processAnswer(player *Player, answer string) error {
	if !player.CanAnswer {
		return fmt.Errorf("player cannot answer")
	}
	player.cancelAnswerTimeout()
	isCorrect := g.CurQuestion.checkAnswer(answer)
	if g.Round == FinalRound {
		return g.processFinalRoundAns(player, isCorrect, answer)
	}
	g.AnsCorrectness = isCorrect
	g.LastAnswer = answer
	g.LastToAnswer = player
	g.setState(RecvVote, &Player{})
	g.messageAllPlayers("Player answered")
	return nil
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
	g.nextQuestion(g.LastToAnswer, isCorrect)
	return nil
}

func (g *Game) processWager(player *Player, wager int) error {
	if !player.CanWager {
		return fmt.Errorf("player cannot wager")
	}
	player.cancelWagerTimeout()
	if min, max, ok := g.validWager(wager, player.Score); !ok {
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
		player.FinalWager = wager
		player.CanWager = false
		g.FinalWagers = append(g.FinalWagers, player.Id)
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

func (g *Game) processProtest(protestByPlayer *Player, protestFor string) error {
	protestForPlayer, err := g.getPlayerById(protestFor)
	if err != nil {
		return err
	}
	if _, ok := protestForPlayer.FinalProtestors[protestByPlayer.Id]; ok {
		return nil
	}
	protestForPlayer.FinalProtestors[protestByPlayer.Id] = true
	if len(protestForPlayer.FinalProtestors) != numPlayers/2+1 {
		_ = protestByPlayer.sendMessage(Response{
			Code:      socket.Ok,
			Message:   "You protested for " + protestForPlayer.Name,
			Game:      g,
			CurPlayer: protestByPlayer,
		})
		return nil
	}
	if protestForPlayer.FinalCorrect {
		protestForPlayer.Score -= 2 * protestForPlayer.FinalWager
	} else {
		protestForPlayer.Score += 2 * protestForPlayer.FinalWager
	}
	protestForPlayer.FinalCorrect = !protestForPlayer.FinalCorrect
	g.setState(PostGame, &Player{})
	g.messageAllPlayers("Final Jeopardy result changed")
	return nil
}

func (g *Game) processFinalRoundAns(player *Player, isCorrect bool, answer string) error {
	player.updateScore(g.CurQuestion.Value, isCorrect, g.Round)
	g.FinalAnswers = append(g.FinalAnswers, player.Id)
	player.CanAnswer = false
	player.FinalAnswer = answer
	player.FinalCorrect = isCorrect
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

func (g *Game) setState(state GameState, player *Player) {
	switch state {
	case BoardIntro:
		for _, p := range g.Players {
			p.updateActions(p.Id == player.Id, false, false, false, false)
		}
		g.startBoardIntroTimeout(player)
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
	if g.Paused {
		g.State = g.PausedState
	}
	g.Paused = false
	state, player := g.State, &Player{}
	if state == PreGame || state == BoardIntro {
		state, player = RecvPick, g.Players[0]
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

func (g *Game) nextQuestion(player *Player, isCorrect bool) {
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
}

func (g *Game) messageAllPlayers(msg string) {
	for _, p := range g.Players {
		_ = p.sendMessage(Response{
			Code:      socket.Ok,
			Message:   msg,
			Game:      g,
			CurPlayer: p,
		})
	}
}

func (g *Game) getPlayerById(id string) (*Player, error) {
	for _, p := range g.Players {
		if p.Id == id {
			return p, nil
		}
	}
	return &Player{}, fmt.Errorf("player not found in game %s", g.Name)
}

func (g *Game) allPlayersReady() bool {
	ready := 0
	for _, p := range g.Players {
		if p.Conn != nil {
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

func (g *Game) lowestPlayer() *Player {
	lowest := g.Players[0]
	for _, p := range g.Players {
		if p.Score < lowest.Score {
			lowest = p
		}
	}
	return lowest
}

func (g *Game) numFinalWagers() int {
	numWagers := 0
	for _, p := range g.Players {
		if p.Score > 0 {
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
