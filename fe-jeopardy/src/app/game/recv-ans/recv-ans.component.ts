import { Component } from '@angular/core';
import { GameStateService } from 'src/app/services/game-state.service';
import { PlayerService } from 'src/app/services/player.service';
import { WebsocketService } from 'src/app/services/websocket.service';

@Component({
    selector: 'app-recv-ans',
    templateUrl: './recv-ans.component.html',
    styleUrls: ['./recv-ans.component.less']
})
export class RecvAnsComponent {
    questionAnswer: string;

    constructor(
        private websocket: WebsocketService,
        protected game: GameStateService,
        protected player: PlayerService,
    ) { }

    handleAnswer() {
        if (this.questionAnswer && this.player.CanAnswer()) {
            this.websocket.Send({
                state: this.game.State(),
                answer: this.questionAnswer,
            })
        }
        this.questionAnswer = '';
    }
}
