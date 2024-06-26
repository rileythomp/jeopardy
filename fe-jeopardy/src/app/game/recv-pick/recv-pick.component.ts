import { Component } from '@angular/core';
import { Question } from 'src/app/model/model';
import { GameStateService } from 'src/app/services/game-state.service';
import { PlayerService } from 'src/app/services/player.service';
import { WebsocketService } from 'src/app/services/websocket.service';

@Component({
	selector: 'app-recv-pick',
	templateUrl: './recv-pick.component.html',
	styleUrls: ['./recv-pick.component.less']
})
export class RecvPickComponent {
	categories: string[];
	questionRows: Question[][];

	constructor(
		private websocket: WebsocketService,
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
			this.websocket.Send({
				state: this.game.State(),
				catIdx: catIdx,
				valIdx: valIdx,
			})
		}
	}
}
