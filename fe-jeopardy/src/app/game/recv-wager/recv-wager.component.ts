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
        private websocket: WebsocketService,
        protected game: GameStateService,
        protected player: PlayerService,
    ) { }

    handleWager() {
        if (this.wagerAmt != null && this.wagerAmt !== '' && this.player.CanWager()) {
            this.websocket.Send({
                state: this.game.State(),
                wager: this.wagerAmt,
            })
        }
        this.wagerAmt = ''
    }
}
