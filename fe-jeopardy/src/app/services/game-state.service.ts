import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';
import { Game, Player, Question, GameState, RoundState } from '../model/model';

@Injectable({
	providedIn: 'root'
})
export class GameStateService {
	private game: Game;
	private gameStateSubject = new Subject<any>();

	constructor() {
		this.game = <Game>{};
	}

	onGameStateChange() {
		return this.gameStateSubject.asObservable();
	}

	updateGameState(newState: Game) {
		this.game = newState;
		this.gameStateSubject.next(this.game);
	}

	Name(): string {
		return this.game.name;
	}

	State(): GameState {
		return this.game.state;
	}

	Players(): Player[] {
		return this.game.players;
	}

	QuestionRows(): Question[][] {
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

	Topics(): string[] {
		let round = this.game.firstRound;
		if (this.game.round == RoundState.SecondRound) {
			round = this.game.secondRound;
		}
		return round.map((topic: {title: string}) => topic.title);
	}

	PickingPlayer(): string {
		return this.game.players.find((player: Player) => player.canPick)?.name ?? '';
	}

	AnsweringPlayer(): string {
		return this.game.players.find((player: Player) => player.canAnswer)?.name ?? '';
	}

	LastToAnswer(): string {
		return this.game.lastToAnswer.name;
	}

	WageringPlayer(): string {
		return this.game.players.find((player: Player) => player.canWager)?.name ?? '';
	}

	LastAnswer(): string {
		return this.game.lastAnswer;
	}

	FinalAnswer(): string {
		return this.game.finalQuestion.answer;
	}

	AnsCorrectness(): boolean {
		return this.game.ansCorrectness;
	}

	IsPaused(): boolean {
		return this.game.paused;
	}

	PreGame(): boolean {
		return this.game.state == GameState.PreGame;
	}

	RecvPick(): boolean {
		return this.game.state == GameState.RecvPick;
	}

	RecvBuzz(): boolean {
		return this.game.state == GameState.RecvBuzz;
	}

	RecvAns(): boolean {
		return this.game.state == GameState.RecvAns;
	}

	RecvVote(): boolean {
		return this.game.state == GameState.RecvVote;
	}

	RecvWager(): boolean {
		return this.game.state == GameState.RecvWager;
	}

	FinalRound(): boolean {
		return this.game.round == RoundState.FinalRound;
	}

	PostGame(): boolean {
		return this.game.state == GameState.PostGame;
	}

	HighestScorers(): string[] {
		const maxScore = Math.max(...this.game.players.map(player => player.score));
		return this.game.players.filter(player => player.score == maxScore).map(player => player.name);
	}

	EndGameMessage(): string {
		if (this.game.state != GameState.PostGame) {
			return '';
		}
		let winners = this.HighestScorers();
		if (winners.length == 1) {
			return `The winner is ${winners[0]}`;
		} else if (winners.length == 2) {
			return `The winners are ${winners[0]} and ${winners[1]}`;
		}
		return `All players have tied`
	}

	CurQuestion(): string {
		return this.game.curQuestion.question;
	}

	CurValue(): number {
		return this.game.curQuestion.value;
	}

	QuestionCanBePicked(topicIdx: number, valIdx: number): boolean {
		let round = this.game.firstRound;
		if (this.game.round == RoundState.SecondRound) {
			round = this.game.secondRound;
		}
		return round[topicIdx].questions[valIdx].canChoose;
	}

	CurQuestionFirstBuzz(): boolean {
		return !this.game.guessedWrong || this.game.guessedWrong.length == 0;
	}
}
