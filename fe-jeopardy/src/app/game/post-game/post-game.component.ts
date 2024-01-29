import { Component } from '@angular/core';
import { GameStateService } from '../../services/game-state.service';
import { Player } from '../../model/model';
import { PlayerService } from '../../services/player.service';
import { WebsocketService } from '../../services/websocket.service';
import { ApiService } from '../../services/api.service';
import { Router } from '@angular/router';

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
	) { }

	canProtestForPlayer(player: Player): boolean {
		return !Object.keys(player.finalProtestors).includes(this.player.Id());
	}

	protestFinalCorrectness(playerId: string) {
		this.websocketService.Send({ protestFor: playerId });
	}

	playAgain() {
		return this.apiService.PlayAgain().subscribe({
			next: (resp: any) => {
				this.player.updatePlayer(resp.player)
			},
			error: (resp: any) => {
				alert(resp.error.message)
			},
		})
	}

	leaveGame() {
		return this.apiService.LeaveGame().subscribe({
			next: (resp: any) => {
				this.router.navigate(['/'])
			},
			error: (resp: any) => {
				alert(resp.error.message)
			},
		})
	}
}
