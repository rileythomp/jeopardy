package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/auth"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/jeopardy"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/socket"
)

type (
	Route struct {
		Method  string
		Path    string
		Handler gin.HandlerFunc
	}

	TokenRequest struct {
		Token string `json:"token,omitempty"`
	}
)

var (
	Routes = []Route{
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/health",
			Handler: CheckHealth,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/version",
			Handler: GetVersion,
		},
		{
			Method:  http.MethodPost,
			Path:    "/jeopardy/games",
			Handler: CreatePrivateGame,
		},
		{
			Method:  http.MethodPut,
			Path:    "/jeopardy/games",
			Handler: JoinPublicGame,
		},
		{
			Method:  http.MethodPut,
			Path:    "/jeopardy/games/:gameCode",
			Handler: JoinGameByCode,
		},
		{
			Method:  http.MethodPut,
			Path:    "/jeopardy/games/bot",
			Handler: AddBot,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/play",
			Handler: PlayGame,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/players/game",
			Handler: GetPlayerGame,
		},
		{
			Method:  http.MethodPost,
			Path:    "/jeopardy/leave",
			Handler: LeaveGame,
		},
		{
			Method:  http.MethodPut,
			Path:    "/jeopardy/play-again",
			Handler: PlayAgain,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/private",
			Handler: GetPrivateGames,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/public",
			Handler: GetPublicGames,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/player-games",
			Handler: GetPlayerGames,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/chat",
			Handler: JoinGameChat,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/analytics",
			Handler: GetAnalytics,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/categories",
			Handler: SearchCategories,
		},
		{
			Method:  http.MethodPut,
			Path:    "/jeopardy/games/start",
			Handler: StartGame,
		},
	}

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == os.Getenv("ALLOW_ORIGIN")
		},
	}
)

const (
	UnexpectedServerErrMsg = "Sorry, there was an unexpected error. Please try again in a few moments. If the issue persists, please file an issue."
	ErrGeneratingJWTMsg    = "Error generating JWT: %s"
	ErrGettingPlayerIdMsg  = "Error getting playerId from token: %s"
	ErrJoiningChatMsg      = "Uh oh, something went wrong when joining the game chat."
	ErrInvalidAuthCredMsg  = "Uh oh, something went wrong: Invalid authentication credentials"
	ErrMalformedReqMsg     = "Uh oh, something went wrong: Malformed request"
)

func GetPlayerGame(c *gin.Context) {
	log.Infof("Received get player game request")

	token := c.Request.Header.Get("Access-Token")
	playerId, err := auth.GetJWTSubject(token)
	if err != nil {
		log.Errorf(ErrGettingPlayerIdMsg, err.Error())
		respondWithError(c, http.StatusForbidden, ErrInvalidAuthCredMsg)
		return
	}

	game, err := jeopardy.GetPlayerGame(playerId)
	if err != nil {
		log.Errorf("Error getting player game: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, "Unable to rejoin game: %s", err.Error())
		return
	}

	c.JSON(http.StatusOK, jeopardy.Response{
		Code:    http.StatusOK,
		Message: "Authorized to get player game",
		Game:    game,
	})
}

func CreatePrivateGame(c *gin.Context) {
	log.Infof("Received create game request")

	var req jeopardy.GameRequest
	if err := parseBody(c.Request.Body, &req); err != nil {
		log.Errorf("Error parsing create request: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, ErrMalformedReqMsg)
		return
	}

	game, playerId, err, code := jeopardy.CreatePrivateGame(req)
	if err != nil {
		log.Errorf("Error creating private game: %s", err.Error())
		if code == socket.BadRequest {
			respondWithError(c, http.StatusBadRequest, "Unable to create private game: %s", err.Error())
		} else {
			respondWithError(c, http.StatusInternalServerError, UnexpectedServerErrMsg)
		}
		return
	}

	jwt, err := auth.GenerateJWT(playerId)
	if err != nil {
		log.Errorf(ErrGeneratingJWTMsg, err.Error())
		respondWithError(c, http.StatusInternalServerError, UnexpectedServerErrMsg)
		return
	}

	c.JSON(http.StatusOK, jeopardy.Response{
		Code:    http.StatusOK,
		Token:   jwt,
		Message: "Authorized to create private game",
		Game:    game,
	})
}

func JoinGameByCode(c *gin.Context) {
	log.Infof("Received private join game request")

	var req jeopardy.GameRequest
	if err := parseBody(c.Request.Body, &req); err != nil {
		log.Errorf("Error parsing join request: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, ErrMalformedReqMsg)
		return
	}

	game, playerId, err := jeopardy.JoinGameByCode(req)
	if err != nil {
		log.Errorf("Error joining game by code: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, "Unable to join game: %s", err.Error())
		return
	}

	jwt, err := auth.GenerateJWT(playerId)
	if err != nil {
		log.Errorf(ErrGeneratingJWTMsg, err.Error())
		respondWithError(c, http.StatusInternalServerError, UnexpectedServerErrMsg)
		return
	}

	c.JSON(http.StatusOK, jeopardy.Response{
		Code:    http.StatusOK,
		Token:   jwt,
		Message: "Authorized to join game by code",
		Game:    game,
	})
}

func JoinPublicGame(c *gin.Context) {
	log.Infof("Received public join game request")

	var req jeopardy.GameRequest
	if err := parseBody(c.Request.Body, &req); err != nil {
		log.Errorf("Error parsing join request: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, ErrMalformedReqMsg)
		return
	}

	game, playerId, err, code := jeopardy.JoinPublicGame(req)
	if err != nil {
		log.Errorf("Error joining public game: %s", err.Error())
		if code == socket.BadRequest {
			respondWithError(c, http.StatusBadRequest, "Unable to join game: %s", err.Error())
		} else {
			respondWithError(c, http.StatusInternalServerError, UnexpectedServerErrMsg)
		}
		return
	}

	jwt, err := auth.GenerateJWT(playerId)
	if err != nil {
		log.Errorf(ErrGeneratingJWTMsg, err.Error())
		respondWithError(c, http.StatusInternalServerError, UnexpectedServerErrMsg)
		return
	}

	c.JSON(http.StatusOK, jeopardy.Response{
		Code:    http.StatusOK,
		Token:   jwt,
		Message: "Authorized to join public game",
		Game:    game,
	})
}

func JoinGameChat(c *gin.Context) {
	log.Infof("Received request to join game chat")

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("Error upgrading connection to WebSocket: %s", err.Error())
		respondWithError(c, http.StatusInternalServerError, ErrJoiningChatMsg)
		return
	}

	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Errorf("Error reading message from WebSocket: %s", err.Error())
		closeConnWithMsg(ws, socket.ServerError, ErrJoiningChatMsg)
		return
	}

	var req TokenRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		log.Errorf("Error parsing chat request: %s", err.Error())
		closeConnWithMsg(ws, socket.BadRequest, ErrMalformedReqMsg)
		return
	}

	playerId, err := auth.GetJWTSubject(req.Token)
	if err != nil {
		log.Errorf(ErrGettingPlayerIdMsg, err.Error())
		closeConnWithMsg(ws, socket.Unauthorized, ErrInvalidAuthCredMsg)
		return
	}

	conn := socket.NewSafeConn(ws)
	err = jeopardy.JoinGameChat(playerId, conn)
	if err != nil {
		log.Errorf("Error joining chat: %s", err.Error())
		closeConnWithMsg(ws, socket.BadRequest, "Unable to join chat: %s", err.Error())
		return
	}
}

func AddBot(c *gin.Context) {
	log.Infof("Received add bot request")

	token := c.Request.Header.Get("Access-Token")
	playerId, err := auth.GetJWTSubject(token)
	if err != nil {
		log.Errorf(ErrGettingPlayerIdMsg, err.Error())
		respondWithError(c, http.StatusForbidden, ErrInvalidAuthCredMsg)
		return
	}

	err = jeopardy.AddBot(playerId)
	if err != nil {
		log.Errorf("Error adding bot to game: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, "Unable to add bot to game: %s", err.Error())
		return
	}
}

func PlayGame(c *gin.Context) {
	log.Infof("Received play request")

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("Error upgrading connection to WebSocket: %s", err.Error())
		respondWithError(c, http.StatusInternalServerError, UnexpectedServerErrMsg)
		return
	}

	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Errorf("Error reading message from WebSocket: %s", err.Error())
		closeConnWithMsg(ws, socket.ServerError, UnexpectedServerErrMsg)
		return
	}

	var req TokenRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		log.Errorf("Error parsing play request: %s", err.Error())
		closeConnWithMsg(ws, socket.BadRequest, ErrMalformedReqMsg)
		return
	}

	playerId, err := auth.GetJWTSubject(req.Token)
	if err != nil {
		log.Errorf(ErrGettingPlayerIdMsg, err.Error())
		closeConnWithMsg(ws, socket.Unauthorized, ErrInvalidAuthCredMsg)
		return
	}

	conn := socket.NewSafeConn(ws)
	err = jeopardy.PlayGame(playerId, conn)
	if err != nil {
		log.Errorf("Error playing game: %s", err.Error())
		closeConnWithMsg(ws, socket.BadRequest, "Unable to play game: %s", err.Error())
		return
	}
}

func PlayAgain(c *gin.Context) {
	log.Infof("Received play again request")

	token := c.Request.Header.Get("Access-Token")
	playerId, err := auth.GetJWTSubject(token)
	if err != nil {
		log.Errorf(ErrGettingPlayerIdMsg, err.Error())
		respondWithError(c, http.StatusForbidden, ErrInvalidAuthCredMsg)
		return
	}

	if err = jeopardy.PlayAgain(playerId); err != nil {
		log.Errorf("Error playing again: %s", err.Error())
		respondWithError(c, socket.BadRequest, "Unable to play again: %s", err.Error())
		return
	}

	c.JSON(http.StatusOK, jeopardy.Response{
		Code:    http.StatusOK,
		Message: "Asked to play again",
	})
}

func LeaveGame(c *gin.Context) {
	log.Infof("Received leave request")

	token := c.Request.Header.Get("Access-Token")
	playerId, err := auth.GetJWTSubject(token)
	if err != nil {
		log.Errorf(ErrGettingPlayerIdMsg, err.Error())
		respondWithError(c, http.StatusForbidden, ErrInvalidAuthCredMsg)
		return
	}

	if err = jeopardy.LeaveGame(playerId); err != nil {
		log.Errorf("Error leaving game: %s", err.Error())
	}

	c.JSON(http.StatusOK, jeopardy.Response{
		Code:    http.StatusOK,
		Message: "Player left the game",
	})
}

func GetAnalytics(c *gin.Context) {
	log.Infof("Received request to get analytics")

	analytics, err := jeopardy.GetAnalytics()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Unable to get analytics")
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func SearchCategories(c *gin.Context) {
	category := c.Query("category")
	rounds := c.Query("rounds")

	categories, err := jeopardy.SearchCategories(category, rounds)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Unable to search categories")
		return
	}

	c.JSON(http.StatusOK, categories)
}

func StartGame(c *gin.Context) {
	log.Infof("Received request to start game")

	token := c.Request.Header.Get("Access-Token")
	playerId, err := auth.GetJWTSubject(token)
	if err != nil {
		log.Errorf(ErrGettingPlayerIdMsg, err.Error())
		respondWithError(c, http.StatusForbidden, ErrInvalidAuthCredMsg)
		return
	}

	err = jeopardy.StartGame(playerId)
	if err != nil {
		log.Errorf("Error starting game: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, "Unable to start game: %s", err.Error())
		return
	}
}

func GetPrivateGames(c *gin.Context) {
	log.Infof("Received request to get private games")
	games := jeopardy.GetPrivateGames()
	c.JSON(http.StatusOK, games)
}

func GetPublicGames(c *gin.Context) {
	log.Infof("Received request to get public games")
	games := jeopardy.GetPublicGames()
	c.JSON(http.StatusOK, games)
}

func GetPlayerGames(c *gin.Context) {
	log.Infof("Received request to get player games")
	playerGames := jeopardy.GetPlayerGames()
	c.JSON(http.StatusOK, playerGames)
}

func CheckHealth(c *gin.Context) {
	log.Infof("Received health check")
	c.String(http.StatusOK, "OK")
}

func GetVersion(c *gin.Context) {
	log.Infof("Received version request")
	info := struct {
		Name    string `json:"name"`
		Domain  string `json:"domain"`
		Version string `json:"version"`
	}{
		Name:    os.Getenv("HEROKU_APP_NAME"),
		Domain:  os.Getenv("HEROKU_APP_DEFAULT_DOMAIN_NAME"),
		Version: os.Getenv("HEROKU_RELEASE_VERSION"),
	}
	c.JSON(http.StatusOK, info)
}

func parseBody(body io.ReadCloser, v any) error {
	msg, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(msg, v)
}

func closeConnWithMsg(conn *websocket.Conn, code int, msg string, args ...any) {
	_ = conn.WriteJSON(jeopardy.Response{Code: code, Message: fmt.Sprintf(msg, args...)})
	_ = conn.Close()
}

func respondWithError(c *gin.Context, code int, msg string, args ...any) {
	c.JSON(code, jeopardy.Response{Code: code, Message: fmt.Sprintf(msg, args...)})
}
