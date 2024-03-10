# Jeopardy

Real-time multi-player Jeopardy game with over 40,000 questions.
Supports public or private games, and includes an in-game chat feature.

## Server

```
$ cd jeopardy-be
$ source .env
$ go mod tidy
$ make run
```

## Client

```
$ cd jeopardy-fe
$ npm install
$ ng serve
```

### TODO

- Improve instructions

- Tournament play

- Allow players to select bot difficulty

- Leaderboards

- Accounts with user stats

- Look into pion for voice chat
