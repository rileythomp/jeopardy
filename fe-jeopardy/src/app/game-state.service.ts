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

	getName(): string {
		return this.game.name;
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

	getLastToAnswer(): string {
		return this.game.lastToAnswer.name;
	}

	getWageringPlayer(): string {
		return this.game.players.find((player: Player) => player.canWager)?.name ?? '';
	}

	getLastAnswer(): string {
		return this.game.lastAnswer;
	}

	getFinalAnswer(): string {
		return this.game.finalQuestion.answer;
	}

	getAnsCorrectness(): boolean {
		return this.game.ansCorrectness;
	}

	isPaused(): boolean {
		return this.game.paused;
	}

	preGame(): boolean {
		return this.game.state == GameState.PreGame;
	}

	recvPick(): boolean {
		return this.game.state == GameState.RecvPick;
	}

	recvBuzz(): boolean {
		return this.game.state == GameState.RecvBuzz;
	}

	recvAns(): boolean {
		return this.game.state == GameState.RecvAns;
	}

	recvVote(): boolean {
		return this.game.state == GameState.RecvVote;
	}

	recvWager(): boolean {
		return this.game.state == GameState.RecvWager;
	}

	finalRound(): boolean {
		return this.game.round == RoundState.FinalRound;
	}

	getHighestScorers(): string[] {
		const maxScore = Math.max(...this.game.players.map(player => player.score));
		return this.game.players.filter(player => player.score == maxScore).map(player => player.name);
	}

	endGameMessage(): string {
		if (this.game.state != GameState.PostGame) {
			return '';
		}
		let winners = this.getHighestScorers();
		if (winners.length == 1) {
			return `The winner is ${winners[0]}`;
		} else if (winners.length == 2) {
			return `The winners are ${winners[0]} and ${winners[1]}`;
		}
		return `All players have tied`
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

	curQuestionFirstBuzz(): boolean {
		return !this.game.guessedWrong || this.game.guessedWrong.length == 0;
	}
}
