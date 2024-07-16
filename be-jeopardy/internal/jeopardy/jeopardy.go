package jeopardy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/socket"
)

type GameRequest struct {
	PlayerName            string        `json:"name"`
	PlayerImg             string        `json:"imgUrl"`
	PlayerEmail           string        `json:"email"`
	Bots                  int           `json:"bots"`
	FullGame              bool          `json:"fullGame"`
	Penalty               bool          `json:"penalty"`
	PickConfig            int           `json:"pickConfig"`
	BuzzConfig            int           `json:"buzzConfig"`
	AnswerConfig          int           `json:"answerConfig"`
	WagerConfig           int           `json:"wagerConfig"`
	FirstRoundCategories  []db.Category `json:"firstRoundCategories"`
	SecondRoundCategories []db.Category `json:"secondRoundCategories"`
}

var GameFull = fmt.Errorf("Game is full")

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

func (g *Game) validateName(name string) error {
	if len(name) < 1 || len(name) > 50 {
		return fmt.Errorf("Invalid player name")
	}
	for _, p := range g.Players {
		if p.name() == name && p.conn() != nil {
			return fmt.Errorf("Sorry, %s is already taken", name)
		}
	}
	return nil
}

func CreatePrivateGame(ctx context.Context, req GameRequest) (*Game, string, error, int) {
	jeopardyDB, err := db.NewJeopardyDB(ctx)
	if err != nil {
		return &Game{}, "", err, socket.ServerError
	}
	config, err := NewConfig(
		req.FullGame, req.Penalty, req.Bots,
		req.PickConfig, req.BuzzConfig, req.AnswerConfig, req.WagerConfig,
		req.FirstRoundCategories, req.SecondRoundCategories,
	)
	if err != nil {
		return &Game{}, "", err, socket.BadRequest
	}
	game, err := NewGame(ctx, jeopardyDB, config)
	if err != nil {
		return &Game{}, "", err, socket.ServerError
	}
	privateGames[game.Name] = game

	if err := game.validateName(req.PlayerName); err != nil {
		return &Game{}, "", err, socket.BadRequest
	}

	imgUrl := req.PlayerImg
	if imgUrl == "" {
		imgUrl = game.nextImg()
	}
	player := NewPlayer(req.PlayerName, imgUrl, req.PlayerEmail)
	game.Players = append(game.Players, player)
	playerGames[player.Id] = game

	for i := 0; i < game.Bots; i++ {
		bot := NewBot(genBotName(), i)
		game.Players = append(game.Players, bot)
		bot.processMessages()
	}

	return game, player.Id, nil, 0
}

func JoinPublicGame(ctx context.Context, req GameRequest) (*Game, string, error, int) {
	var game *Game
	for _, g := range publicGames {
		if len(g.Players) < maxPlayers && g.validateName(req.PlayerName) == nil {
			game = g
			break
		}
	}
	if game == nil {
		jeopardyDB, err := db.NewJeopardyDB(ctx)
		if err != nil {
			return &Game{}, "", err, socket.ServerError
		}
		config, err := NewConfig(
			req.FullGame, req.Penalty, req.Bots,
			req.PickConfig, req.BuzzConfig, req.AnswerConfig, req.WagerConfig,
			req.FirstRoundCategories, req.SecondRoundCategories,
		)
		if err != nil {
			return &Game{}, "", err, socket.BadRequest
		}
		game, err = NewGame(ctx, jeopardyDB, config)
		if err != nil {
			return &Game{}, "", err, socket.ServerError
		}
		publicGames[game.Name] = game
	}

	if err := game.validateName(req.PlayerName); err != nil {
		return &Game{}, "", err, socket.BadRequest
	}
	imgUrl := req.PlayerImg
	if imgUrl == "" {
		imgUrl = game.nextImg()
	}
	player := NewPlayer(req.PlayerName, imgUrl, req.PlayerEmail)
	game.Players = append(game.Players, player)
	playerGames[player.Id] = game

	return game, player.Id, nil, socket.Ok
}

func findGame(joinCode string) *Game {
	for _, g := range publicGames {
		if strings.EqualFold(g.Name, joinCode) || strings.EqualFold(g.Code, joinCode) {
			return g
		}
	}
	for _, g := range privateGames {
		if strings.EqualFold(g.Name, joinCode) || strings.EqualFold(g.Code, joinCode) {
			return g
		}
	}
	return nil
}

func JoinGameByCode(req GameRequest, joinCode string) (*Game, string, error) {
	game := findGame(joinCode)
	if game == nil {
		return &Game{}, "", fmt.Errorf("Game not found")
	}

	if err := game.validateName(req.PlayerName); err != nil {
		return &Game{}, "", err
	}

	var player GamePlayer
	for _, p := range game.Players {
		if p.conn() == nil {
			delete(playerGames, p.id())
			player = p
			player.setId(uuid.New().String())
			player.setName(req.PlayerName)
			imgUrl := req.PlayerImg
			if imgUrl == "" {
				imgUrl = game.nextImg()
			}
			player.setImg(imgUrl)
			break
		}
	}
	if player == nil {
		if len(game.Players) >= maxPlayers {
			return &Game{}, "", GameFull
		}
		imgUrl := req.PlayerImg
		if imgUrl == "" {
			imgUrl = game.nextImg()
		}
		player = NewPlayer(req.PlayerName, imgUrl, req.PlayerEmail)
		game.Players = append(game.Players, player)
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

	var bot *Bot
	for i, p := range game.Players {
		if p.conn() == nil {
			delete(playerGames, p.id())
			bot = NewBot(genBotName(), game.numBots())
			bot.copyState(p)
			game.Players[i] = bot
			break
		}
	}
	if bot == nil {
		if len(game.Players) >= maxPlayers {
			return GameFull
		}
		bot = NewBot(genBotName(), game.numBots())
		game.Players = append(game.Players, bot)
	}

	bot.processMessages()

	game.messageAllPlayers("Waiting to start")

	return nil
}

func PlayGame(playerId, gameName string, conn SafeConn) error {
	game, err := GetPlayerGame(playerId)
	if err != nil {
		return err
	}

	if game.Name != gameName {
		log.Errorf("Player's current game name '%s' and given game name '%s' do not match", game.Name, gameName)
		return fmt.Errorf("Game names does not match")
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

	game.messageAllPlayers("Waiting to start")

	return nil
}

func StartGame(playerId string) error {
	game, err := GetPlayerGame(playerId)
	if err != nil {
		return err
	}

	if game.Disconnected {
		game.Disconnected = false
	}
	if game.Paused {
		game.resumeGame()
	} else {
		game.startRound(game.Players[0])
	}
	game.messageAllPlayers("We are ready to play")

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

var searchDB *db.JeopardyDB
var analyticsDB *db.JeopardyDB
var supabase *db.SupabaseDB

func init() {
	ctx := context.Background()
	var err error
	searchDB, err = db.NewJeopardyDB(ctx)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err.Error())
	}
	analyticsDB, err = db.NewJeopardyDB(ctx)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err.Error())
	}
	supabase, err = db.NewSupabaseDB(ctx)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err.Error())
	}
}

func SearchCategories(ctx context.Context, category, rounds string) ([]db.Category, error) {
	if category == "" {
		return []db.Category{}, nil
	}
	start := ""
	if len(category) > 2 {
		start = "%"
	}
	secondRound := 2
	if rounds == "first" {
		secondRound = 1
	}
	categories, err := searchDB.SearchCategories(ctx, strings.ToLower(category), start, secondRound)
	if err != nil {
		log.Errorf("Error searching categories: %s", err.Error())
		return nil, err
	}
	return categories, nil
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
	g.jeopardyDB.Close()
	delete(publicGames, g.Name)
	delete(privateGames, g.Name)
	for _, p := range g.Players {
		delete(playerGames, p.id())
	}
}
