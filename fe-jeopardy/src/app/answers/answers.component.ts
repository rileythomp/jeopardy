import { Component } from '@angular/core';
import { GameStateService } from '../services/game-state.service';
import { PlayerService } from '../services/player.service';
import { WebsocketService } from '../services/websocket.service';

@Component({
	selector: 'app-answers',
	templateUrl: './answers.component.html',
	styleUrls: ['./answers.component.less']
})
export class AnswersComponent {

	constructor(
		protected game: GameStateService,
		protected player: PlayerService,
		private websocket: WebsocketService,
	) { }

	disputeQuestion() {
		this.websocket.Send({
			state: this.game.State(),
			initDispute: true,
		})
	}
}
