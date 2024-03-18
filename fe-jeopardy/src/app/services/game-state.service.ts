import { Injectable } from '@angular/core'
import { Subject } from 'rxjs'
import { Game, Player, Question, GameState, RoundState, Answer } from '../model/model'

@Injectable({
	providedIn: 'root'
})
export class GameStateService {
	private game: Game
	private gameStateSubject = new Subject<any>()

	constructor() {
		this.game = <Game>{}
	}

	onGameStateChange() {
		return this.gameStateSubject.asObservable()
	}

	updateGameState(newState: Game) {
		this.game = newState
		this.gameStateSubject.next(this.game)
	}

	Name(): string {
		return this.game.name
	}

	State(): GameState {
		return this.game.state
	}

	Players(): Player[] {
		if (!this.game.players) {
			this.game.players = []
		}
		return this.game.players
	}

	QuestionRows(): Question[][] {
		let firstRow = []
		let secondRow = []
		let thirdRow = []
		let fourthRow = []
		let fifthRow = []
		let round = this.game.firstRound
		if (this.game.round == RoundState.SecondRound) {
			round = this.game.secondRound
		}
		for (let category of round) {
			firstRow.push(category.questions[0])
			secondRow.push(category.questions[1])
			thirdRow.push(category.questions[2])
			fourthRow.push(category.questions[3])
			fifthRow.push(category.questions[4])
		}
		return [firstRow, secondRow, thirdRow, fourthRow, fifthRow]
	}

	Categories(): string[] {
		let round = this.game.firstRound
		if (this.game.round == RoundState.SecondRound) {
			round = this.game.secondRound
		}
		return round.map((category: { title: string }) => category.title)
	}

	PickingPlayer(): string {
		return this.game.players.find((player: Player) => player.canPick)?.name ?? ''
	}

	AnsweringPlayer(): string {
		return this.game.players.find((player: Player) => player.canAnswer)?.name ?? ''
	}

	WageringPlayer(): string {
		return this.game.players.find((player: Player) => player.canWager)?.name ?? ''
	}

	AnsCorrectness(): boolean {
		return this.game.ansCorrectness
	}

	IsPaused(): boolean {
		return this.game.paused
	}

	PreGame(): boolean {
		return this.game.state == GameState.PreGame
	}

	BoardIntro(): boolean {
		return this.game.state == GameState.BoardIntro
	}

	RecvPick(): boolean {
		return this.game.state == GameState.RecvPick
	}

	RecvBuzz(): boolean {
		return this.game.state == GameState.RecvBuzz
	}

	RecvAns(): boolean {
		return this.game.state == GameState.RecvAns
	}

	RecvWager(): boolean {
		return this.game.state == GameState.RecvWager
	}

	FinalRound(): boolean {
		return this.game.round == RoundState.FinalRound
	}

	PostGame(): boolean {
		return this.game.state == GameState.PostGame
	}

	HighestScorers(): string[] {
		const maxScore = Math.max(...this.game.players.map(player => player.score))
		return this.game.players.filter(player => player.score == maxScore).map(player => player.name)
	}

	CurCategory(): string {
		return this.game.curQuestion.category
	}

	CurComments(): string {
		return this.game.curQuestion.comments
	}

	CurQuestion(): string {
		return this.game.curQuestion.question
	}

	OfficialAnswer(): string {
		return this.game.officialAnswer
	}

	CurValue(): number {
		return this.game.curQuestion.value
	}

	QuestionCanBePicked(catIdx: number, valIdx: number): boolean {
		let round = this.game.firstRound
		if (this.game.round == RoundState.SecondRound) {
			round = this.game.secondRound
		}
		return round[catIdx].questions[valIdx].canChoose
	}

	CurQuestionFirstBuzz(): boolean {
		return !this.game.guessedWrong || this.game.guessedWrong.length == 0
	}

	StartBuzzCountdown(): boolean {
		return this.game.startBuzzCountdown
	}

	StartFinalWagerCountdown(): boolean {
		return this.game.startFinalWagerCountdown
	}

	StartFinalAnswerCountdown(): boolean {
		return this.game.startFinalAnswerCountdown
	}

	BlockBuzz(block: boolean): void {
		this.game.buzzBlocked = block
	}

	BuzzBlocked(): boolean {
		return this.game.buzzBlocked
	}

	BuzzDelay(): number {
		return Math.min(Math.max(Math.floor(this.game.curQuestion.question.split(' ').length / 5), 3), 9)
	}

	InDispute(): boolean {
		return this.game.state == GameState.RecvDispute
	}

	CanInitDispute(playerId: string): boolean {
		let answers = this.game.curQuestion.answers
		for (let i = 0; i < answers?.length; i++) {
			let ans = answers[i]
			if (ans.player.id == playerId && !ans.hasDisputed && !ans.correct && ans.answer != 'answer-timeout') {
				return true
			}
			if (ans.overturned) {
				return false
			}
		}
		return false
	}

	DisputerName(): string {
		return this.game.curQuestion.curDisputed.player.name
	}

	DisputerAnswer(): string {
		return this.game.curQuestion.curDisputed.answer
	}

	CurAnswerer(): string {
		return this.game.curQuestion.curAns.player.name
	}

	CurAnswer(): string {
		return this.game.curQuestion.curAns.answer
	}

	CurQuestionAnswers(): Answer[] {
		return this.game.curQuestion.answers
	}

	AdjustedAnswers(): Answer[] {
		let adjustments: Answer[] = []
		let add = false
		let answers = this.game.curQuestion.answers
		for (let i = 0; i < answers?.length; i++) {
			let ans = answers[i]
			if (ans.player.name == this.DisputerName()) {
				add = true
			} else if (add) {
				adjustments.push(ans)
			}
			if (ans.overturned) {
				break
			}
		}
		return adjustments;
	}

	Disconnected(): boolean {
		return this.game.disconnected
	}

	Penalty(): boolean {
		return this.game.penalty
	}

	PickTimeout(): number {
		return this.game.pickTimeout
	}

	BuzzTimeout(): number {
		return this.game.buzzTimeout
	}

	AnswerTimeout(): number {
		return this.game.answerTimeout
	}

	WagerTimeout(): number {
		return this.game.wagerTimeout
	}

	FinalAnswerTimeout(): number {
		return this.game.finalAnswerTimeout
	}

	FinalWagerTimeout(): number {
		return this.game.finalWagerTimeout
	}
}
