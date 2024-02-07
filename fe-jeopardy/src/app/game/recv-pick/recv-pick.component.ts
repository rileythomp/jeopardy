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
	@Input() countdownSeconds: number;
	categories: string[];
	questionRows: Question[][];

	constructor(
		private websocketService: WebsocketService,
		protected game: GameStateService,
		protected player: PlayerService,
	) {
		this.categories = this.game.Categories()
		this.questionRows = this.game.QuestionRows()
	}

	changeCursor(canChoose: boolean, pointer: string) {
		if (canChoose && this.player.CanPick()) {
			document.body.style.cursor = pointer
		}
	}

	handlePick(catIdx: number, valIdx: number) {
		if (this.player.CanPick() && this.game.QuestionCanBePicked(catIdx, valIdx)) {
			document.body.style.cursor = 'default'
			this.websocketService.Send({
				catIdx: catIdx,
				valIdx: valIdx,
			})
		}
	}

}
