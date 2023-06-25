import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';
import { Game, Player, Question, GameState, RoundState } from './model/model';

@Injectable({
	providedIn: 'root'
})
export class GameStateService {
	private game: Game;
	private gameStateSubject = new Subject<any>();

	constructor() { }

	onGameStateChange() {
		return this.gameStateSubject.asObservable();
	}

	updateGameState(newState: Game) {
		this.game = newState;
		this.gameStateSubject.next(this.game);
	}

	getGame() {
		return this.game;
	}

	getGameState(): GameState {
		return this.game.state;
	}

	getPlayers(): Player[] {
		return this.game.players;
	}

	readyToPlay(): boolean {
		return this.game.players.length == 3;
	}

	getQuestionRows(): Question[][] {
		let firstRow = [];
		let secondRow = [];
		let thirdRow = [];
		// let fourthRow = [];
		// let fifthRow = [];
		let round = this.game.firstRound;
		if (this.game.round == RoundState.SecondRound) {
			round = this.game.secondRound;
		}
		for (let topic of round) {
			firstRow.push(topic.questions[0]);
			secondRow.push(topic.questions[1]);
			thirdRow.push(topic.questions[2]);
			// fourthRow.push(topic.questions[3]);
			// fifthRow.push(topic.questions[4]);
		}
		return [firstRow, secondRow, thirdRow];
		// return [firstRow, secondRow, thirdRow, fourthRow, fifthRow];
	}

	getTitles(): string[] {
		let round = this.game.firstRound;
		if (this.game.round == RoundState.SecondRound) {
			round = this.game.secondRound;
		}
		return round.map((topic: {title: string}) => topic.title);
	}

	getPickingPlayer(): string {
		return this.game.players.find((player: Player) => player.canPick)?.name ?? '';
	}

	getAnsweringPlayer(): string {
		return this.game.players.find((player: Player) => player.canAnswer)?.name ?? '';
	}

	getWageringPlayer(): string {
		return this.game.players.find((player: Player) => player.canWager)?.name ?? '';
	}

	recvingPick(): boolean {
		return this.game.state == GameState.RecvPick;
	}

	recvingBuzz(): boolean {
		return this.game.state == GameState.RecvBuzz;
	}

	recvingAns(): boolean {
		return this.game.state == GameState.RecvAns;
	}

	recvingWager(): boolean {
		return this.game.state == GameState.RecvWager;
	}

	finalRound(): boolean {
		return this.game.round == RoundState.FinalRound;
	}

	getWinner(): string {
		let max = 0;
		let winner = '';
		for (let player of this.game.players) {
			if (player.score > max) {
				max = player.score;
				winner = player.name;
			}
		}
		return winner;
	}

	gameOver(): boolean {
		return this.game.state == GameState.PostGame;
	}

	curQuestion(): string {
		return this.game.curQuestion.question;
	}

	curValue(): number {
		return this.game.curQuestion.value;
	}

	questionCanBePicked(topicIdx: number, valIdx: number): boolean {
		let round = this.game.firstRound;
		if (this.game.round == RoundState.SecondRound) {
			round = this.game.secondRound;
		}
		return round[topicIdx].questions[valIdx].canChoose;
	}
}
