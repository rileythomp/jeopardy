package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

	GameRequest struct {
		PlayerName string `json:"playerName"`
		GameCode   string `json:"gameCode"`
	}

	PlayRequest struct {
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
			Method:  http.MethodGet,
			Path:    "/jeopardy/play",
			Handler: PlayGame,
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
			Method:  http.MethodPost,
			Path:    "/jeopardy/leave",
			Handler: LeaveGame,
		},
		{
			Method:  http.MethodPost,
			Path:    "/jeopardy/play-again",
			Handler: PlayAgain,
		},
	}

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
			// TODO: SET THIS PROPERLY
			// return r.Header.Get("Origin") == "http://localhost:4200"
		},
	}
)

func CreatePrivateGame(c *gin.Context) {
	log.Infof("Received create game request")

	var req GameRequest
	if err := parseBody(c.Request.Body, &req); err != nil {
		log.Errorf("Error parsing create request: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, "Error parsing create request")
		return
	}

	game, playerId, err := jeopardy.CreatePrivateGame(req.PlayerName)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error creating game: %s", err.Error())
		return
	}

	jwt, err := auth.GenerateJWT(playerId)
	if err != nil {
		log.Errorf("Error generating JWT: %s", err.Error())
		respondWithError(c, http.StatusInternalServerError, "Error generating JWT")
		return
	}

	c.JSON(http.StatusOK, jeopardy.Response{
		Code:    http.StatusOK,
		Token:   jwt,
		Message: "Authorized to create game",
		Game:    game,
	})
}

func JoinGameByCode(c *gin.Context) {
	log.Infof("Received private join game request")

	var req GameRequest
	if err := parseBody(c.Request.Body, &req); err != nil {
		log.Errorf("Error parsing join request: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, "Error parsing join request")
		return
	}

	game, playerId, err := jeopardy.JoinGameByCode(req.PlayerName, req.GameCode)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error joining game: %s", err.Error())
		return
	}

	jwt, err := auth.GenerateJWT(playerId)
	if err != nil {
		log.Errorf("Error generating JWT: %s", err.Error())
		respondWithError(c, http.StatusInternalServerError, "Error generating JWT")
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

	var req GameRequest
	if err := parseBody(c.Request.Body, &req); err != nil {
		log.Errorf("Error parsing join request: %s", err.Error())
		respondWithError(c, http.StatusBadRequest, "Error parsing join request")
		return
	}

	game, playerId, err := jeopardy.JoinPublicGame(req.PlayerName)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error joining game: %s", err.Error())
		return
	}

	jwt, err := auth.GenerateJWT(playerId)
	if err != nil {
		log.Errorf("Error generating JWT: %s", err.Error())
		respondWithError(c, http.StatusInternalServerError, "Error generating JWT")
		return
	}

	c.JSON(http.StatusOK, jeopardy.Response{
		Code:    http.StatusOK,
		Token:   jwt,
		Message: "Authorized to join game",
		Game:    game,
	})
}

func PlayGame(c *gin.Context) {
	log.Infof("Received play request")

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("Error upgrading connection to WebSocket: %s", err.Error())
		respondWithError(c, http.StatusInternalServerError, "Error playing game")
		return
	}

	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Errorf("Error reading message from WebSocket: %s", err.Error())
		closeConnWithMsg(ws, http.StatusInternalServerError, "Error reading message WebSocket")
		return
	}

	var req PlayRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		log.Errorf("Error parsing play request: %s", err.Error())
		closeConnWithMsg(ws, http.StatusBadRequest, "Error parsing play request")
		return
	}

	playerId, err := auth.GetJWTSubject(req.Token)
	if err != nil {
		log.Errorf("Error getting playerId from token: %s", err.Error())
		closeConnWithMsg(ws, http.StatusForbidden, "Error getting playerId from token")
		return
	}

	conn := socket.NewSafeConn(ws)
	err = jeopardy.PlayGame(playerId, conn)
	if err != nil {
		log.Errorf("Error during game: %s", err.Error())
		closeConnWithMsg(ws, http.StatusInternalServerError, "Error during game: %s", err.Error())
		return
	}
}

func PlayAgain(c *gin.Context) {
	log.Infof("Received play again request")

	c.JSON(http.StatusOK, jeopardy.Response{Message: "PLAYING AGAIN"})
}

func LeaveGame(c *gin.Context) {
	log.Infof("Received leave request")

	c.JSON(http.StatusOK, jeopardy.Response{Message: "LEAVING GAME"})
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

func CheckHealth(c *gin.Context) {
	log.Infof("Received health check")
	c.String(http.StatusOK, "OK")
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
	conn.Close()
}

func respondWithError(c *gin.Context, code int, msg string, args ...any) {
	c.JSON(code, jeopardy.Response{Code: code, Message: fmt.Sprintf(msg, args...)})
}
