package jeopardy

import (
	"fmt"

	"github.com/nwtgck/go-fakelish"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
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

func genGameCode() string {
	return fakelish.GenerateFakeWord(7, 7)
}

func CreatePrivateGame(playerName string) (*Game, string, error) {
	game, err := NewGame(genGameCode())
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
	var game *Game
	for _, g := range publicGames {
		if len(g.Players) < numPlayers {
			game = g
			break
		}
	}
	if game == nil {
		var err error
		game, err = NewGame(genGameCode())
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
	game, ok := publicGames[gameCode]
	if !ok {
		game, ok = privateGames[gameCode]
		if !ok {
			log.Errorf("Game %s not found", gameCode)
			return &Game{}, "", fmt.Errorf("Game %s not found", gameCode)
		}
	}

	var player *Player
	for _, p := range game.Players {
		if p.Conn == nil {
			player = p
			player.Name = playerName
		}
	}
	if player == nil {
		player = NewPlayer(playerName)
		game.Players = append(game.Players, player)
	}
	playerGames[player.Id] = game

	return game, player.Id, nil
}

func PlayGame(playerId string, conn SafeConn) error {
	game := getPlayerGame(playerId)
	if game == nil {
		return fmt.Errorf("no game found for player id: %s", playerId)
	}

	player := game.getPlayerById(playerId)
	if player == nil {
		return fmt.Errorf("no player found for player id")
	}
	if player.Conn != nil {
		return fmt.Errorf("player already playing")
	}
	player.Conn = conn

	msg := "Waiting for more players"
	if game.allPlayersReady() {
		game.startGame()
		msg = "We are ready to play"
	}

	player.sendPings()
	player.processMessages(game.msgChan, game.stopChan)

	// TODO: HANDLE THIS ERROR
	_ = game.messageAllPlayers(msg)
	return nil
}
