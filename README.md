# Jeopardy

Real-time multi-player Jeopardy game with over 100,000 questions.

Play here: https://playjeopardy.netlify.app

<details>
  <summary>Screenshots</summary>
  
![home](imgs/home.png)

![search](imgs/search.png)

![config](imgs/config.png)

</details>

### Features

- Game configuration

  - Play against other people or against bots
  - Choose the categories you want to play with
  - Play 1 or 2 round games
  - Play with or without penalties for incorrect answers
  - Play in public or private games

- Game analytics

  - Total number of games played
  - Average score after each round
  - Buzz-in rate for each round
  - Correctness rate for each round

- In-game chat

- Allows for players to pause the game

- Handles players disconnecting and rejoining the game

### Development

Server

```
$ cd jeopardy-be
$ go mod tidy
$ docker compose up -d postgres
$ source .env
$ make run
```

Client

```
$ cd jeopardy-fe
$ npm install
$ ng serve
```

### TODO

- Tournament play

- Allow players to import their own questions

- Allow players to select bot difficulty

- Leaderboards

- Accounts with user stats

- Look into pion for voice chat
