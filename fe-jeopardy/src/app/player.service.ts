import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';
import { Player } from './model/model';

@Injectable({
	providedIn: 'root'
})
export class PlayerService {
	private player: Player;
	private playerSubject = new Subject<any>();

	constructor() { }

	onPlayerChange() {
		return this.playerSubject.asObservable();
	}

	updatePlayer(newPlayer: Player) {
		this.player = newPlayer;
		this.playerSubject.next(this.player);
	}

	getPlayer(): Player {
		return this.player;
	}

	getName(): string {
		return this.player.name;
	}

	canPick(): boolean {
		return this.player.canPick;
	}

	canBuzz(): boolean {
		return this.player.canBuzz;
	}

	canAnswer(): boolean {
		return this.player.canAnswer;
	}

	canWager(): boolean {
		return this.player.canWager;
	}

	canConfirmAns(): boolean {
		return this.player.canConfirmAns;
	}
}