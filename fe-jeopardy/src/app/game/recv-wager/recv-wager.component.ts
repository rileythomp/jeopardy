import { Component } from '@angular/core';
import { GameStateService } from '../../services/game-state.service';
import { PlayerService } from '../../services/player.service';
import { WebsocketService } from '../../services/websocket.service';

@Component({
    selector: 'app-recv-wager',
    templateUrl: './recv-wager.component.html',
    styleUrls: ['./recv-wager.component.less']
})
export class RecvWagerComponent {
    wagerAmt: string

    constructor(
        private websocketService: WebsocketService,
        protected game: GameStateService,
        protected player: PlayerService,
    ) { }

    handleWager() {
        if (this.wagerAmt != null && this.wagerAmt !== '' && this.player.CanWager()) {
            this.websocketService.Send({
                wager: this.wagerAmt,
            })
        }
        this.wagerAmt = ''
    }
}
