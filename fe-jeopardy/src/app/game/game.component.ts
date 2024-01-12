import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { GameStateService } from '../game-state.service';
import { WebsocketService } from '../websocket.service';
import { PlayerService } from '../player.service';
import { JwtService } from '../jwt.service';
import { Question, GameState, Ping } from '../model/model';

// const  pickTimeout = 10
// const  buzzTimeout = 10
// const  defaultAnsTimeout = 10
// const  dailyDoubleAnsTimeout = 10
// const  finalJeopardyAnsTimeout = 10
// const  voteTimeout = 10
// const  dailyDoubleWagerTimeout = 10
// const  finalJeopardyWagerTimeout = 10
// const  buzzDelay = 2000/2

const pickTimeout = 10
const buzzTimeout = 2
const defaultAnsTimeout = 10
const dailyDoubleAnsTimeout = 10
const finalJeopardyAnsTimeout = 10
const voteTimeout = 2
const dailyDoubleWagerTimeout = 10
const finalJeopardyWagerTimeout = 10
const buzzDelay = 2000 / 2

@Component({
	selector: 'app-game',
	templateUrl: './game.component.html',
	styleUrls: ['./game.component.less'],
})
export class GameComponent implements OnInit {
	private jwt: string;
	private countdownInterval: any;
	protected countdownSeconds: number;
	protected questionAnswer: string;
	protected wagerAmt: string;
	protected lobbyMessage: string;
	protected questionRows: Question[][];
	protected topics: string[];

	constructor(
		private router: Router,
		private websocketService: WebsocketService,
		private jwtService: JwtService,
		protected game: GameStateService,
		protected player: PlayerService,
	) { }

	ngOnInit(): void {
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt;
		});

		this.websocketService.Connect('play');

		this.websocketService.OnOpen(() => {
			let playReq = {
				token: this.jwt,
			}
			this.websocketService.Send(playReq);
		})

		this.websocketService.OnMessage((event: { data: string; }) => {
			let resp = JSON.parse(event.data);

			if (resp.code != 200) {
				// TODO: REPLACE WITH MODAL
				alert(resp.message);
				if (resp.code == 500) {
					this.router.navigate(['/join']);
				}
				return
			}

			if (resp.message == Ping) {
				return
			}

			console.log(resp);

			this.game.updateGameState(resp.game);
			this.player.updatePlayer(resp.curPlayer);
			this.lobbyMessage = resp.message;
			this.topics = this.game.Topics();
			this.questionRows = this.game.QuestionRows();

			if (this.game.IsPaused()) {
				this.countdownSeconds = 0;
				clearInterval(this.countdownInterval);
				// TODO: REPLACE WITH MODAL
				alert(`${resp.message}, will resume when 3 players are ready`);
				return
			}

			switch (this.game.State()) {
				case GameState.PreGame:
				case GameState.PostGame:
					break
				case GameState.RecvBuzz:
					if (this.game.CurQuestionFirstBuzz()) {
						this.player.BlockBuzz(true)
						setTimeout(() => {
							this.player.BlockBuzz(false)
							if (this.player.CanBuzz()) {
								this.startCountdownTimer(buzzTimeout - buzzDelay / 1000);
							}
						}, buzzDelay);
					} else {
						if (this.player.CanBuzz()) {
							this.startCountdownTimer(buzzTimeout);
						}
					}
					break;
				case GameState.RecvAns:
					if (this.player.CanAnswer()) {
						this.startCountdownTimer(defaultAnsTimeout);
					}
					break;
				case GameState.RecvPick:
					if (this.player.CanPick()) {
						this.startCountdownTimer(pickTimeout);
					}
					break;
				case GameState.RecvVote:
					if (this.player.CanVote()) {
						this.startCountdownTimer(voteTimeout);
					}
					break
				case GameState.RecvWager:
					if (this.player.CanWager()) {
						this.startCountdownTimer(dailyDoubleWagerTimeout);
					}
					break;
				default:
					// TODO: REPLACE WITH MODAL
					alert('Unable to update game, redirecting to home page');
					this.router.navigate(['/join']);
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
}
