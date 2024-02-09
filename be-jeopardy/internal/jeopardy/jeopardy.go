package jeopardy

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/socket"
)

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

func CreatePrivateGame(playerName string) (*Game, string, error) {
	if len(playerName) < 1 || len(playerName) > 20 {
		return &Game{}, "", fmt.Errorf("player name must be between 1 and 20 characters")
	}

	questionDB, err := db.NewQuestionDB()
	if err != nil {
		log.Errorf("Error connecting to database: %s", err.Error())
		return &Game{}, "", fmt.Errorf("Unexpected error creating private game")
	}
	game, err := NewGame(questionDB)
	if err != nil {
		log.Errorf("Error creating game: %s", err.Error())
		return &Game{}, "", err
	}
	privateGames[game.Name] = game

	player := NewPlayer(playerName)
	game.Players = append(game.Players, player)
	playerGames[player.Id] = game

	return game, player.Id, nil
}

func JoinPublicGame(playerName string) (*Game, string, error) {
	if len(playerName) < 1 || len(playerName) > 20 {
		return &Game{}, "", fmt.Errorf("player name must be between 1 and 20 characters")
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
			log.Errorf("Error connecting to database: %s", err.Error())
			return &Game{}, "", fmt.Errorf("Unexpected error creating private game")
		}
		game, err = NewGame(questionDB)
		if err != nil {
			log.Errorf("Error creating game: %s", err.Error())
			return &Game{}, "", err
		}
		publicGames[game.Name] = game
	}

	player := NewPlayer(playerName)
	game.Players = append(game.Players, player)
	playerGames[player.Id] = game

	return game, player.Id, nil
}

func JoinGameByCode(playerName, gameCode string) (*Game, string, error) {
	if len(playerName) < 1 || len(playerName) > 20 {
		return &Game{}, "", fmt.Errorf("player name must be between 1 and 20 characters")
	}

	game, ok := publicGames[gameCode]
	if !ok {
		game, ok = privateGames[gameCode]
		if !ok {
			log.Errorf("Game %s not found", gameCode)
			return &Game{}, "", fmt.Errorf("Game %s not found", gameCode)
		}
	}

	var player *Player
	if len(game.Players) < numPlayers {
		player = NewPlayer(playerName)
		game.Players = append(game.Players, player)
	} else {
		for _, p := range game.Players {
			if p.Conn == nil {
				delete(playerGames, p.Id)
				player = p
				player.Id = uuid.New().String()
				player.Name = playerName
				break
			}
		}
	}
	if player == nil {
		return &Game{}, "", fmt.Errorf("game %s is full", gameCode)
	}
	playerGames[player.Id] = game

	return game, player.Id, nil
}

func GetPlayerGame(playerId string) (*Game, error) {
	game, ok := playerGames[playerId]
	if !ok {
		return nil, fmt.Errorf("no game found for player with id %s", playerId)
	}
	return game, nil
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
	if player.Conn != nil {
		return fmt.Errorf("player already playing")
	}
	player.Conn = conn

	msg := "Waiting for more players"
	if game.allPlayersReady() {
		if game.Paused {
			game.startGame()
		} else {
			game.setState(BoardIntro, game.Players[0])
		}
		msg = "We are ready to play"
	}

	player.sendPings()
	player.processMessages(game.msgChan, game.pauseChan)

	game.messageAllPlayers(msg)
	return nil
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

	game.pauseChan <- player

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

	player.PlayAgain = true

	restartGame := true
	for _, p := range game.Players {
		if !p.PlayAgain || p.Conn == nil {
			restartGame = false
		}
	}
	if restartGame {
		game.restartChan <- true
		return nil
	}

	for _, p := range game.Players {
		msg := fmt.Sprintf("%s wants to play again", player.Name)
		if p.Id == player.Id {
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
