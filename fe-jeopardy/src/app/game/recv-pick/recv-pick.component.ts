import { Component, Input } from '@angular/core';
import { GameState, Question } from 'src/app/model/model';
import { PlayerService } from 'src/app/services/player.service';
import { WebsocketService } from 'src/app/services/websocket.service';
import { GameStateService } from 'src/app/services/game-state.service';

@Component({
    selector: 'app-recv-pick',
    templateUrl: './recv-pick.component.html',
    styleUrls: ['./recv-pick.component.less']
})
export class RecvPickComponent {
    @Input() topics: string[];
    @Input() questionRows: Question[][];
    @Input() countdownSeconds: number;

    constructor(
        private websocketService: WebsocketService,
        protected game: GameStateService,
        protected player: PlayerService,
    ) { }

    highlightQuestion(event: any, color: string) {
		if (event.target.style.backgroundColor == 'lightpink') {
			return
		}
		if (this.player.CanPick()) {
			event.target.style.backgroundColor = color;
		}
	}

    handlePick(topicIdx: number, valIdx: number) {
		if (this.player.CanPick() && this.game.QuestionCanBePicked(topicIdx, valIdx)) {
			this.websocketService.Send({
				topicIdx: topicIdx,
				valIdx: valIdx,
			})
		}
	}

}
