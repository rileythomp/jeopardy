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

	Player(): Player {
		return this.player;
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


	SetCanBuzz(canBuzz: boolean): void {
		this.player.canBuzz = canBuzz;
		this.updatePlayer(this.player);
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

	Score(): number {
		return this.player.score;
	}

	Conn(): any {
		return this.player.conn;
	}

	SetCanDispute(canDispute: boolean): void {
		this.player.canDispute = canDispute;
		this.updatePlayer(this.player);
	}
}