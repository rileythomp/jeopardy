import { Component } from '@angular/core';
import { GameStateService } from 'src/app/services/game-state.service';
import { PlayerService } from 'src/app/services/player.service';
import { WebsocketService } from 'src/app/services/websocket.service';

@Component({
    selector: 'app-recv-buzz',
    templateUrl: './recv-buzz.component.html',
    styleUrls: ['./recv-buzz.component.less']
})
export class RecvBuzzComponent {
    constructor(
        private websocket: WebsocketService,
        protected game: GameStateService,
        protected player: PlayerService
    ) { }

    handleBuzz(pass: boolean) {
        if (this.player.CanBuzz()) {
            this.websocket.Send({
                state: this.game.State(),
                isPass: pass,
            })
            this.player.SetCanBuzz(false)
        }
    }
}
