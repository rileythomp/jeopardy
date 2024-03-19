import { Component } from '@angular/core';
import { GameStateService } from '../services/game-state.service';
import { PlayerService } from '../services/player.service';
import { WebsocketService } from '../services/websocket.service';

@Component({
	selector: 'app-dispute',
	templateUrl: './dispute.component.html',
	styleUrls: ['./dispute.component.less']
})
export class DisputeComponent {
	constructor(
		protected game: GameStateService,
		protected player: PlayerService,
		private websocket: WebsocketService
	) { }

	disputeQuestion(dispute: boolean) {
		if (this.player.CanDispute()) {
			this.websocket.Send({
				state: this.game.State(),
				dispute: dispute,
			})
			this.player.SetCanDispute(false)
		}
	}
}
