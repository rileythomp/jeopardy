import { Component, OnInit } from '@angular/core';
import { GameStateService } from '../game-state.service';
import { WebsocketService } from '../websocket.service';
import { PlayerService } from '../player.service';
import { JwtService } from '../jwt.service';
import { Player, Question, GameState, Ping } from '../model/model';

const  pickTimeout = 10
const  buzzTimeout = 10
const  defaultAnsTimeout = 10
const  dailyDoubleAnsTimeout = 10
const  finalJeopardyAnsTimeout = 10
const  voteTimeout = 10
const  dailyDoubleWagerTimeout = 10
const  finalJeopardyWagerTimeout = 10
const  buzzDelay = 2000/2

// const pickTimeout               = 2
// const buzzTimeout               = 2
// const defaultAnsTimeout         = 10
// const dailyDoubleAnsTimeout     = 10
// const finalJeopardyAnsTimeout   = 10
// const voteTimeout               = 2
// const dailyDoubleWagerTimeout   = 10
// const finalJeopardyWagerTimeout = 10
// const  buzzDelay = 2000/2

@Component({
	selector: 'app-game',
	templateUrl: './game.component.html',
	styleUrls: ['./game.component.less'],
})
export class GameComponent implements OnInit {
	players: Player[];
	titles: string[];
	questionRows: Question[][];
	questionAnswer: string;
	wagerAmt: string;
	countdownSeconds: number;
	countdownInterval: any;

	constructor(
		private websocketService: WebsocketService,
		private jwtService: JwtService,
		protected gameState: GameStateService,
		protected player: PlayerService,
	) { }

	ngOnInit(): void {
		this.players = this.gameState.getPlayers();
		this.titles = this.gameState.getTitles();
		this.questionRows = this.gameState.getQuestionRows();

		if (this.gameState.isPaused()) {
			alert('Game is paused, will resume when 3 players are ready');
		} else {
			this.initCountdownTimer(this.gameState.getGameState());
		}

		this.websocketService.onmessage((event: { data: string; }) => {
			let resp = JSON.parse(event.data);
			if (resp.code != 200) {
				// TODO: REPLACE ALERTS WITH MODALS
				// alert(resp.message)
				console.log('restarting wager timeout')
				this.startCountdownTimer(dailyDoubleWagerTimeout)
				return
			}

			if (resp.message == Ping) {
				return
			}

			if (resp.game.paused) {
				this.countdownSeconds = 0;
				clearInterval(this.countdownInterval);
				alert(`${resp.message}, will resume when 3 players are ready`);
				return
			}

			switch (resp.game.state) {
				case GameState.PreGame: 
					console.log('a player has left the game');
					console.log(resp);

					this.gameState.updateGameState(resp.game);
					this.player.updatePlayer(resp.curPlayer);

					break

				case GameState.RecvBuzz:
					console.log('show the question, accept a buzz');
					console.log(resp);

					this.gameState.updateGameState(resp.game);
					this.player.updatePlayer(resp.curPlayer);

					this.players = this.gameState.getPlayers();
					this.questionRows = this.gameState.getQuestionRows();

					if (this.gameState.curQuestionFirstBuzz()) {
						this.player.blockBuzz(true)
						setTimeout(() => {
							this.player.blockBuzz(false)
							if (this.player.canBuzz()) {
								this.startCountdownTimer(buzzTimeout - buzzDelay/1000);
							}
						}, buzzDelay);
					} else {
						if (this.player.canBuzz()) {
							this.startCountdownTimer(buzzTimeout);
						}
					}
					break;

				case GameState.RecvAns:
					console.log('alert of buzz, accept an answer');
					console.log(resp);

					this.gameState.updateGameState(resp.game);
					this.player.updatePlayer(resp.curPlayer);

					if (this.player.canAnswer()) {
						this.startCountdownTimer(defaultAnsTimeout);
					}

					break;

				case GameState.RecvPick:
					console.log('show the board, accept a pick');
					console.log(resp);

					this.gameState.updateGameState(resp.game);
					this.player.updatePlayer(resp.curPlayer);

					this.players = this.gameState.getPlayers();
					this.titles = this.gameState.getTitles();
					this.questionRows = this.gameState.getQuestionRows();

					if (this.player.canPick()) {
						this.startCountdownTimer(pickTimeout);
					}

					break;

				case GameState.RecvVote:
					console.log("show the answers correctness, accept a vote");
					console.log(resp);

					this.gameState.updateGameState(resp.game);
					this.player.updatePlayer(resp.curPlayer);

					if (this.player.canVote()) {
						this.startCountdownTimer(voteTimeout);
					}

					break

				case GameState.RecvWager:
					console.log('show the question, accept a wager');
					console.log(resp);

					this.gameState.updateGameState(resp.game);
					this.player.updatePlayer(resp.curPlayer);

					this.players = this.gameState.getPlayers();

					if (this.player.canWager()) {
						this.startCountdownTimer(dailyDoubleWagerTimeout);
					}

					break;

				case GameState.PostGame:
					console.log('show who won the game');
					console.log(resp);

					this.gameState.updateGameState(resp.game);
					this.player.updatePlayer(resp.curPlayer);

					this.players = this.gameState.getPlayers();
					break;

				default:
					alert('Unable to update game');
					break;
			}
		})
	}

	startCountdownTimer(seconds: number) {
		clearInterval(this.countdownInterval)
		this.countdownSeconds = seconds;
		this.countdownInterval = setInterval(() => {
			this.countdownSeconds -= 1;
			if (this.countdownSeconds <= 0) {
				clearInterval(this.countdownInterval);
			}
		}, 1000);
	}

	initCountdownTimer(gameState: GameState) {
		if (gameState == GameState.RecvPick && this.player.canPick()) {
			this.startCountdownTimer(pickTimeout);
		}
		else if (gameState == GameState.RecvBuzz && this.player.canBuzz()) {
			this.startCountdownTimer(buzzTimeout);
		}
		else if (gameState == GameState.RecvAns && this.player.canAnswer()) {
			this.startCountdownTimer(defaultAnsTimeout);
		}
		else if (gameState == GameState.RecvVote && this.player.canVote()) {
			this.startCountdownTimer(voteTimeout);
		}
		else if (gameState == GameState.RecvWager && this.player.canWager()) {
			this.startCountdownTimer(dailyDoubleWagerTimeout);
		}
	}

	highlightQuestion(event: any, color: string) {
		if (event.target.style.backgroundColor == 'lightpink') {
			return
		}
		if (this.player.canPick()) {
			event.target.style.backgroundColor = color;
		}
	}

	handleQuestionPick(topicIdx: number, valIdx: number) {
		if (this.player.canPick() && this.gameState.questionCanBePicked(topicIdx, valIdx)) {
			this.websocketService.send({
				"token": this.jwtService.getJwt(),
				"topicIdx": topicIdx,
				"valIdx": valIdx,
			})
		}
	}

	handleBuzz(pass: boolean) {
		if (this.player.canBuzz()) {
			this.websocketService.send({
				"token": this.jwtService.getJwt(),
				"isPass": pass,
			})
		}
	}

	handleAnswer() {
		if (this.player.canAnswer()) {
			this.websocketService.send({
				"token": this.jwtService.getJwt(),
				"answer": this.questionAnswer,
			})
		}
		this.questionAnswer = '';
	}

	handleVote(confirm: boolean) {
		if (this.player.canVote()) {
			this.websocketService.send({
				"token": this.jwtService.getJwt(),
				"confirm": confirm,
			})
		}
	}

	handleWager() {
		if (this.player.canWager()) {
			this.websocketService.send({
				"token": this.jwtService.getJwt(),
				"wager": this.wagerAmt,
			})
		}
		this.wagerAmt = '';
	}

	protestFinalCorrectness(playerId: string) {
		this.websocketService.send({
			"token": this.jwtService.getJwt(),
			"protestFor": playerId,
		})
	}

	canProtestForPlayer(player: Player): boolean {
		return !Object.keys(player.finalProtestors).includes(this.player.getPlayer().id);
	}
}
