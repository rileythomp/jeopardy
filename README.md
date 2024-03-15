# Jeopardy

Real-time multi-player Jeopardy game with over 100,000 questions.

Features:

- Game configuration

  - Play against other people or against bots
  - Play in public or private games
  - Play 1 or 2 round games
  - Play with or without penalties for incorrect answers

- Game analytics

  - Total number of games played
  - Average score after each round
  - Buzz-in rate for each round
  - Correctness rate for each round

- In-game chat

- Allows for players to pause the game

- Handles players disconnecting and rejoining the game

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

- Tournament play

- Allow players to search for and select their own categories

- Allow players to import their own questions

- Allow players to select bot difficulty

- Leaderboards

- Accounts with user stats

- Look into pion for voice chat
