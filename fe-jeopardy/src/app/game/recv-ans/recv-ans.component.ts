import { Component, Input } from '@angular/core';
import { GameStateService } from 'src/app/game-state.service';
import { PlayerService } from 'src/app/player.service';
import { WebsocketService } from 'src/app/websocket.service';

@Component({
    selector: 'app-recv-ans',
    templateUrl: './recv-ans.component.html',
    styleUrls: ['./recv-ans.component.less']
})
export class RecvAnsComponent {
    @Input() countdownSeconds: number;
    questionAnswer: string;

    constructor(
        private websocketService: WebsocketService,
        protected game: GameStateService,
        protected player: PlayerService,
    ) { }

    handleAnswer() {
		if (this.player.CanAnswer()) {
			this.websocketService.Send({
				answer: this.questionAnswer,
			})
		}
		this.questionAnswer = '';
	}
}
