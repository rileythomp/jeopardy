package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/auth"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/handlers"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	if err := auth.SetJWTKeys(); err != nil {
		log.Fatalf("Failed to set JWT keys: %s", err)
	}

	r := gin.Default()
	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatalf("Failed to set trusted proxies: %s", err)
	}
	r.Use(cors.Default())

	for _, route := range handlers.Routes {
		r.Handle(route.Method, route.Path, route.Handler)
	}

	port := os.Getenv("PORT")
	addr := flag.String("addr", ":"+port, "http service address")
	log.Fatal(r.Run(*addr))
}
