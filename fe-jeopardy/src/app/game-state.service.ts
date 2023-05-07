import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';

@Injectable({
	providedIn: 'root'
})
export class GameStateService {
	private gameState: any;
	private gameStateSubject = new Subject<any>();

	constructor() { }

	onGameStateChange() {
		return this.gameStateSubject.asObservable();
	}

	updateGameState(newState: {}) {
		this.gameState = newState;
		this.gameStateSubject.next(this.gameState);
	}

	getGameState() {
		return this.gameState;
	}

	playerNames(): string[] {
		return Object.values(this.gameState.players).map((player: any) => player.name);
	}

	readyToPlay(): boolean {
		return this.playerNames().length == 3;
	}
}
