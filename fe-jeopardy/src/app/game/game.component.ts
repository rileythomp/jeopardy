import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { GameStateService } from '../game-state.service';
import { WebsocketService } from '../websocket.service';
import { PlayerService } from '../player.service';
import { JwtService } from '../jwt.service';
import { ApiService } from '../api.service';
import { Player, Question, GameState, Ping } from '../model/model';

// const  pickTimeout = 10
// const  buzzTimeout = 10
// const  defaultAnsTimeout = 10
// const  dailyDoubleAnsTimeout = 10
// const  finalJeopardyAnsTimeout = 10
// const  voteTimeout = 10
// const  dailyDoubleWagerTimeout = 10
// const  finalJeopardyWagerTimeout = 10
// const  buzzDelay = 2000/2

const pickTimeout               = 2
const buzzTimeout               = 2
const defaultAnsTimeout         = 10
const dailyDoubleAnsTimeout     = 10
const finalJeopardyAnsTimeout   = 10
const voteTimeout               = 2
const dailyDoubleWagerTimeout   = 10
const finalJeopardyWagerTimeout = 10
const  buzzDelay = 2000/2

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
		private router: Router,
		private websocketService: WebsocketService,
		private jwtService: JwtService,
		protected gameState: GameStateService,
		protected player: PlayerService,
		private apiService: ApiService,
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
				alert(resp.message);
				if (resp.code == 500) {
					this.router.navigate(['/join']);
				}
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

			console.log(resp);

			if (!(resp.game.state in GameState)) {
				alert('Unable to update game');
				return
			}

			this.gameState.updateGameState(resp.game);
			this.player.updatePlayer(resp.curPlayer);

			switch (resp.game.state) {
				case GameState.PreGame: 
					break
				case GameState.RecvBuzz:
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
					if (this.player.canAnswer()) {
						this.startCountdownTimer(defaultAnsTimeout);
					}
					break;
				case GameState.RecvPick:
					this.players = this.gameState.getPlayers();
					this.titles = this.gameState.getTitles();
					this.questionRows = this.gameState.getQuestionRows();
					if (this.player.canPick()) {
						this.startCountdownTimer(pickTimeout);
					}
					break;
				case GameState.RecvVote:
					if (this.player.canVote()) {
						this.startCountdownTimer(voteTimeout);
					}
					break
				case GameState.RecvWager:
					this.players = this.gameState.getPlayers();
					if (this.player.canWager()) {
						this.startCountdownTimer(dailyDoubleWagerTimeout);
					}
					break;
				case GameState.PostGame:
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

	handlePick(topicIdx: number, valIdx: number) {
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

	playAgain() {
		return this.apiService.playAgain({"hello": "world"}).subscribe({
			next: (resp: any) => {
				console.log('playing again', resp)
			},
			error: (err: any) => {
				console.log('Error playing again', err)
				alert('Error playing again')
			},
		})
	}

	leaveGame() {
		return this.apiService.leaveGame({"hello": "world"}).subscribe({
			next: (resp: any) => {
				console.log('left game', resp)
			},
			error: (err: any) => {
				console.log('Error leaving game', err)
				alert('Error leaving game')
			},
		})
	}
}
