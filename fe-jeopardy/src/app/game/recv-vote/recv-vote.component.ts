import { Component } from '@angular/core';
import { GameStateService } from '../../services/game-state.service';
import { PlayerService } from '../../services/player.service';
import { WebsocketService } from '../../services/websocket.service';

@Component({
    selector: 'app-recv-vote',
    templateUrl: './recv-vote.component.html',
    styleUrls: ['./recv-vote.component.less']
})
export class RecvVoteComponent {
    constructor(
        private websocketService: WebsocketService,
        protected game: GameStateService,
        protected player: PlayerService,
    ) { }

    handleVote(confirm: boolean) {
        if (this.player.CanVote()) {
            this.websocketService.Send({
                state: this.game.State(),
                confirm: confirm,
            })
        }
    }
}
