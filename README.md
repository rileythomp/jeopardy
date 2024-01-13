# Jeopardy

Real-time multi-player Jeopardy game.

## Server

```
$ cd jeopardy-be
$ go mod tidy
$ make run
```

## Client

```
$ cd jeopardy-fe
$ ng serve
```

### TODO 

* Make public games a map, provide links for games

* Handle end game 

* Use correct timeouts on FE

* Hanle blank answers, timeouts and closed connections on the UI more gracefully 

* Improve UI 

* Fetch questions from DB
