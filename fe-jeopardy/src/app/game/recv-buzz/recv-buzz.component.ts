import { Component, Input } from '@angular/core';
import { GameStateService } from 'src/app/services/game-state.service';
import { PlayerService } from 'src/app/services/player.service';
import { WebsocketService } from 'src/app/services/websocket.service';

@Component({
    selector: 'app-recv-buzz',
    templateUrl: './recv-buzz.component.html',
    styleUrls: ['./recv-buzz.component.less']
})
export class RecvBuzzComponent {
    @Input() countdownSeconds: number;

    constructor(
        private websocketService: WebsocketService,
        protected game: GameStateService,
        protected player: PlayerService
    ) { }

    handleBuzz(pass: boolean) {
		if (this.player.CanBuzz()) {
			this.websocketService.Send({
				isPass: pass,
			})
		}
	}
}
