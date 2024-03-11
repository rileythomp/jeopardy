import { Injectable } from '@angular/core';
import { GameStateService } from './game-state.service';

@Injectable({
	providedIn: 'root'
})
export class ModalService {
	private messageTimeout: NodeJS.Timeout
	gameMessage: string
	message: boolean
	instructions: boolean
	analytics: boolean

	constructor(private game: GameStateService) { }

	displayMessage(msg: string) {
		this.instructions = false
		this.analytics = false
		clearTimeout(this.messageTimeout)
		this.gameMessage = msg
		this.message = true
		this.messageTimeout = setTimeout(() => {
			this.message = false
		}, 10000)
	}

	displayDispute() {
		this.instructions = false
		this.analytics = false
	}

	displayInstructions() {
		if (this.game.InDispute()) {
			return
		}
		this.analytics = false
		this.instructions = true
	}

	displayAnalytics() {
		if (this.game.InDispute()) {
			return
		}
		this.instructions = false
		this.analytics = true
	}

	hideGameMessage() {
		this.message = false
	}

	hideInstructions() {
		this.instructions = false
	}

	hideAnalytics() {
		this.analytics = false
	}

	getGameMessage(): string {
		return this.gameMessage
	}

	showMessage(): boolean {
		return this.message
	}

	showInstructions(): boolean {
		return this.instructions
	}

	showAnalytics(): boolean {
		return this.analytics
	}
}
