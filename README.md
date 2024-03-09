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

- Track and show game analytics

- Allow players to select 1 or 2 round

- Remove alerts

- Allow players to select bot difficulty

- Accounts with user stats

- Leaderboards

- Tournament play

- Look into pion for voice chat
