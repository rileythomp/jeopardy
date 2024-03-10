import { Component } from '@angular/core';
import { GameStateService } from '../../services/game-state.service';
import { Player, ServerUnavailableMsg } from '../../model/model';
import { PlayerService } from '../../services/player.service';
import { WebsocketService } from '../../services/websocket.service';
import { ApiService } from '../../services/api.service';
import { Router } from '@angular/router';
import { ModalService } from 'src/app/services/modal.service';

@Component({
	selector: 'app-post-game',
	templateUrl: './post-game.component.html',
	styleUrls: ['./post-game.component.less']
})
export class PostGameComponent {

	constructor(
		private router: Router,
		private apiService: ApiService,
		private websocketService: WebsocketService,
		protected game: GameStateService,
		protected player: PlayerService,
		private modal: ModalService,
	) { }

	canProtestForPlayer(player: Player): boolean {
		return !Object.keys(player.finalProtestors).includes(this.player.Id());
	}

	protestFinalCorrectness(playerId: string) {
		this.websocketService.Send({
			state: this.game.State(),
			protestFor: playerId
		});
	}

	playAgain() {
		return this.apiService.PlayAgain().subscribe({
			next: (resp: any) => {
				console.log(resp)
			},
			error: (err: any) => {
				let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg;
				this.modal.displayMessage(msg)
			},
		})
	}

	leaveGame() {
		return this.apiService.LeaveGame().subscribe({
			next: (resp: any) => {
				this.router.navigate(['/'])
			},
			error: (err: any) => {
				let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg;
				console.log(msg)
				this.router.navigate(['/'])
			},
		})
	}
}
