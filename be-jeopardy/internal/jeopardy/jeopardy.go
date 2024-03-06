package jeopardy

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/socket"
)

type GameRequest struct {
	PlayerName string `json:"playerName"`
	GameCode   string `json:"gameCode"`
	Bots       int    `json:"bots"`
}

var (
	privateGames = map[string]*Game{}
	publicGames  = map[string]*Game{}
	playerGames  = map[string]*Game{}
)

func GetPublicGames() map[string]*Game {
	return publicGames
}

func GetPrivateGames() map[string]*Game {
	return privateGames
}

func GetPlayerGames() map[string]string {
	playerGameNames := map[string]string{}
	for playerId, game := range playerGames {
		playerGameNames[playerId] = game.Name
	}
	return playerGameNames
}

func validateName(name string) error {
	if len(name) < 1 || len(name) > 20 {
		return fmt.Errorf("Player name must be between 1 and 20 characters")
	}
	return nil
}

func CreatePrivateGame(req GameRequest) (*Game, string, error, int) {
	if err := validateName(req.PlayerName); err != nil {
		return &Game{}, "", err, socket.BadRequest
	}

	questionDB, err := db.NewQuestionDB()
	if err != nil {
		return &Game{}, "", err, socket.ServerError
	}
	game, err := NewGame(questionDB)
	if err != nil {
		return &Game{}, "", err, socket.ServerError
	}
	privateGames[game.Name] = game

	player := NewPlayer(req.PlayerName)
	game.Players = append(game.Players, player)
	playerGames[player.Id] = game

	for i := 0; i < req.Bots; i++ {
		bot := NewBot(genGameCode())
		game.Players = append(game.Players, bot)
		bot.processMessages()
	}

	return game, player.Id, nil, 0
}

func JoinPublicGame(playerName string) (*Game, string, error, int) {
	if err := validateName(playerName); err != nil {
		return &Game{}, "", err, socket.BadRequest
	}

	var game *Game
	for _, g := range publicGames {
		if len(g.Players) < numPlayers {
			game = g
			break
		}
	}
	if game == nil {
		questionDB, err := db.NewQuestionDB()
		if err != nil {
			return &Game{}, "", err, socket.ServerError
		}
		game, err = NewGame(questionDB)
		if err != nil {
			return &Game{}, "", err, socket.ServerError
		}
		publicGames[game.Name] = game
	}

	player := NewPlayer(playerName)
	game.Players = append(game.Players, player)
	playerGames[player.Id] = game

	return game, player.Id, nil, socket.Ok
}

func JoinGameByCode(playerName, gameCode string) (*Game, string, error) {
	if err := validateName(playerName); err != nil {
		return &Game{}, "", err
	}

	game, ok := publicGames[gameCode]
	if !ok {
		game, ok = privateGames[gameCode]
		if !ok {
			return &Game{}, "", fmt.Errorf("Game %s not found", gameCode)
		}
	}

	var player GamePlayer
	if len(game.Players) < numPlayers {
		player = NewPlayer(playerName)
		game.Players = append(game.Players, player)
	} else {
		for _, p := range game.Players {
			if p.conn() == nil {
				delete(playerGames, p.id())
				player = p
				player.setId(uuid.New().String())
				player.setName(playerName)
				break
			}
		}
	}
	if player == nil {
		return &Game{}, "", fmt.Errorf("Game %s is full", gameCode)
	}
	playerGames[player.id()] = game

	return game, player.id(), nil
}

func GetPlayerGame(playerId string) (*Game, error) {
	game, ok := playerGames[playerId]
	if !ok {
		return nil, fmt.Errorf("No game found for player")
	}
	return game, nil
}

func AddBot(playerId string) error {
	game, err := GetPlayerGame(playerId)
	if err != nil {
		return err
	}

	if len(game.Players) >= numPlayers {
		return fmt.Errorf("Game is full")
	}

	bot := NewBot(genGameCode())
	game.Players = append(game.Players, bot)
	bot.processMessages()

	game.handlePlayerJoined(bot)

	return nil
}

func PlayGame(playerId string, conn SafeConn) error {
	game, err := GetPlayerGame(playerId)
	if err != nil {
		return err
	}

	player, err := game.getPlayerById(playerId)
	if err != nil {
		return err
	}
	if player.conn() != nil {
		return fmt.Errorf("Player already playing")
	}
	player.setConn(conn)
	player.sendPings()
	player.readMessages(game.msgChan, game.disconnectChan)

	game.handlePlayerJoined(player)

	return nil
}

func (g *Game) handlePlayerJoined(player GamePlayer) {
	msg := "Waiting for more players"
	if g.allPlayersReady() {
		if g.Paused {
			g.startGame()
		} else {
			g.setState(BoardIntro, g.Players[0])
			// g.setState(RecvPick, g.Players[0])
		}
		msg = "We are ready to play"
	}

	g.messageAllPlayers(msg)
}

func LeaveGame(playerId string) error {
	game, err := GetPlayerGame(playerId)
	if err != nil {
		return err
	}

	player, err := game.getPlayerById(playerId)
	if err != nil {
		return err
	}

	game.disconnectChan <- player

	return nil
}

func PlayAgain(playerId string) error {
	game, err := GetPlayerGame(playerId)
	if err != nil {
		return err
	}

	player, err := game.getPlayerById(playerId)
	if err != nil {
		return err
	}

	player.setPlayAgain(true)

	restartGame := true
	for _, p := range game.Players {
		if !p.playAgain() || p.conn() == nil {
			restartGame = false
		}
	}
	if restartGame {
		game.restartChan <- true
		return nil
	}

	for _, p := range game.Players {
		msg := fmt.Sprintf("%s wants to play again", player.name())
		if p.id() == player.id() {
			msg = "Waiting for all other players to play again"
		}
		_ = p.sendMessage(Response{
			Code:      socket.Info,
			Message:   msg,
			Game:      game,
			CurPlayer: p,
		})
	}
	return nil
}

func CleanUpGames() {
	log.Infof("Performing game cleanup")
	for _, game := range publicGames {
		if game.Paused && time.Since(game.PausedAt) > time.Hour {
			log.Infof("Game %s has been paused for over an hour, removing it", game.Name)
			removeGame(game)
		}
	}
	for _, game := range privateGames {
		if game.Paused && time.Since(game.PausedAt) > time.Hour {
			log.Infof("Game %s has been paused for over an hour, removing it", game.Name)
			removeGame(game)
		}
	}
}

func removeGame(g *Game) {
	if err := g.questionDB.Close(); err != nil {
		log.Errorf("Error closing question db: %s", err.Error())
	}
	delete(publicGames, g.Name)
	delete(privateGames, g.Name)
	for _, p := range g.Players {
		delete(playerGames, p.id())
	}
}
