import { Component, OnInit } from '@angular/core';
import { GameStateService } from '../game-state.service';
import { WebsocketService } from '../websocket.service';
import { PlayerService } from '../player.service';
import { JwtService } from '../jwt.service';
import { Player, Question, GameState } from '../model/model';

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

		this.websocketService.onmessage((event: { data: string; }) => {
			let resp = JSON.parse(event.data);
			if (resp.game.state == GameState.RecvBuzz) {
				console.log('show the question, accept a buzz');
				console.log(resp);

				this.gameState.updateGameState(resp.game);
				this.player.updatePlayer(resp.curPlayer);

				this.players = this.gameState.getPlayers();
				this.questionRows = this.gameState.getQuestionRows();
			}
			else if (resp.game.state == GameState.RecvAns) {
				console.log('alert of buzz, accept an answer');
				console.log(resp);

				this.gameState.updateGameState(resp.game);
				this.player.updatePlayer(resp.curPlayer);
			}
			else if (resp.game.state == GameState.RecvPick) {
				console.log('show the board, accept a pick');
				console.log(resp);

				this.gameState.updateGameState(resp.game);
				this.player.updatePlayer(resp.curPlayer);

				this.players = this.gameState.getPlayers();
				this.titles = this.gameState.getTitles();
				this.questionRows = this.gameState.getQuestionRows();
			}
			else if (resp.game.state == GameState.RecvWager) {
				console.log('show the question, accept a wager');
				console.log(resp);

				this.gameState.updateGameState(resp.game);
				this.player.updatePlayer(resp.curPlayer);

				this.players = this.gameState.getPlayers();
			}
			else if (resp.game.state == GameState.PostGame) {
				console.log('show who won the game');
				console.log(resp);

				this.gameState.updateGameState(resp.game);
				this.player.updatePlayer(resp.curPlayer);

				this.players = this.gameState.getPlayers();
			}
			else {
				alert('Unable to update game');
			}
		})
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

	handleBuzz() {
		if (this.player.canBuzz()) {
			this.websocketService.send({
				"token": this.jwtService.getJwt(),
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

	handleWager() {
		if (this.player.canWager()) {
			this.websocketService.send({
				"token": this.jwtService.getJwt(),
				"wager": this.wagerAmt,
			})
		}
		this.wagerAmt = '';
	}
}
