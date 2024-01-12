import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';
import { Player } from './model/model';

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

	BlockBuzz(block: boolean): void {
		this.player.buzzBlocked = block;
	}

	CanPick(): boolean {
		return this.player.canPick;
	}

	CanBuzz(): boolean {
		return this.player.canBuzz && !this.player.buzzBlocked;
	}

	CanAnswer(): boolean {
		return this.player.canAnswer;
	}

	CanWager(): boolean {
		return this.player.canWager;
	}

	CanVote(): boolean {
		return this.player.canVote;
	}
}