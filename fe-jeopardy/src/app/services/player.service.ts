import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';
import { Player } from '../model/model';
import { GameStateService } from './game-state.service';

@Injectable({
	providedIn: 'root'
})
export class PlayerService {
	private player: Player;
	private playerSubject = new Subject<any>();

	constructor() {
		this.player = <Player>{};
	}

	onPlayerChange() {
		return this.playerSubject.asObservable();
	}

	updatePlayer(newPlayer: Player) {
		this.player = newPlayer;
		this.playerSubject.next(this.player);
	}

	Id(): string {
		return this.player.id;
	}

	Name(): string {
		return this.player.name;
	}

	CanPick(): boolean {
		return this.player.canPick;
	}

	CanBuzz(): boolean {
		return this.player.canBuzz;
	}

	CanAnswer(): boolean {
		return this.player.canAnswer;
	}

	CanWager(): boolean {
		return this.player.canWager;
	}

	CanDispute(): boolean {
		return this.player.canDispute;
	}

	PlayAgain(): boolean {
		return this.player.playAgain;
	}

	FinalWager(): number {
		return this.player.finalWager;
	}
}