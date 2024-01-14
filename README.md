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

* Handle players disconnecting when in the start state

* Only show players on the UI if their connection is not nil

* Handle end game 

* Use correct timeouts on FE

* Hanle blank answers, timeouts and closed connections on the UI more gracefully 

* Improve UI 

* Fetch questions from DB
