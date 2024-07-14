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
	config: boolean
	register: boolean
	login: boolean

	constructor(private game: GameStateService) { }

	displayMessage(msg: string) {
		this.login = false
		this.register = false
		this.config = false
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
		this.login = false
		this.register = false
		this.config = false
		this.instructions = false
		this.analytics = false
	}

	displayInstructions() {
		if (this.game.InDispute()) {
			return
		}
		this.login = false
		this.register = false
		this.config = false
		this.analytics = false
		this.instructions = true
	}

	displayAnalytics() {
		if (this.game.InDispute()) {
			return
		}
		this.login = false
		this.register = false
		this.config = false
		this.instructions = false
		this.analytics = true
	}

	displayConfig() {
		if (this.game.InDispute()) {
			return
		}
		this.login = false
		this.register = false
		this.instructions = false
		this.analytics = false
		this.config = true
	}

	displayRegister() {
		if (this.game.InDispute()) {
			return
		}
		this.login = false
		this.config = false
		this.instructions = false
		this.analytics = false
		this.register = true
	}

	displayLogin() {
		if (this.game.InDispute()) {
			return
		}
		this.config = false
		this.instructions = false
		this.analytics = false
		this.register = false
		this.login = true
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

	hideConfig() {
		this.config = false
	}

	hideRegister() {
		this.register = false
	}

	hideLogin() {
		this.login = false
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

	showConfig(): boolean {
		return this.config
	}

	showRegister(): boolean {
		return this.register
	}

	showLogin(): boolean {
		return this.login
	}
}
