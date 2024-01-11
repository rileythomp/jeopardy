package handlers

import (
	"encoding/json"
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

	JoinRequest struct {
		PlayerName string `json:"playerName"`
		GameName   string `json:"gameName"`
		Private    bool   `json:"private"`
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
			Path:    "/jeopardy/join",
			Handler: JoinGame,
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

func JoinGame(c *gin.Context) {
	log.Infof("Received join request")

	var req JoinRequest
	if err := parseBody(c.Request.Body, &req); err != nil {
		log.Errorf("Error parsing join request: %s", err.Error())
		respondWithError(c, "Error parsing join request", http.StatusBadRequest)
		return
	}

	game, playerId, err := jeopardy.JoinGame(req.PlayerName, req.GameName, req.Private)
	if err != nil {
		respondWithError(c, "Error joining game", http.StatusInternalServerError)
		return
	}

	jwt, err := auth.GenerateJWT(playerId)
	if err != nil {
		log.Errorf("Error generating JWT: %s", err.Error())
		respondWithError(c, "Error generating JWT", http.StatusInternalServerError)
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
		respondWithError(c, "Error playing game", http.StatusInternalServerError)
		return
	}

	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Errorf("Error reading message from WebSocket: %s", err.Error())
		closeConnWithMsg(ws, "Error reading message WebSocket", http.StatusInternalServerError)
		return
	}

	var req PlayRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		log.Errorf("Error parsing play request: %s", err.Error())
		closeConnWithMsg(ws, "Error parsing play request", http.StatusBadRequest)
		return
	}

	playerId, err := auth.GetJWTSubject(req.Token)
	if err != nil {
		log.Errorf("Error getting playerId from token: %s", err.Error())
		closeConnWithMsg(ws, "Error getting playerId from token", http.StatusForbidden)
		return
	}

	conn := socket.NewSafeConn(ws)
	err = jeopardy.PlayGame(playerId, conn)
	if err != nil {
		log.Errorf("Error during game: %s", err.Error())
		closeConnWithMsg(ws, "Error during game", http.StatusInternalServerError)
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

func closeConnWithMsg(conn *websocket.Conn, msg string, code int) {
	_ = conn.WriteJSON(jeopardy.Response{Code: code, Message: msg})
	conn.Close()
}

func respondWithError(c *gin.Context, msg string, code int) {
	c.JSON(code, jeopardy.Response{Code: code, Message: msg})
}
