import { Component, Input } from '@angular/core';
import { GameStateService } from '../../services/game-state.service';
import { environment } from '../../../environments/environment';
import { ApiService } from '../../services/api.service';

@Component({
	selector: 'app-pre-game',
	templateUrl: './pre-game.component.html',
	styleUrls: ['./pre-game.component.less']
})
export class PreGameComponent {
	@Input() gameName: string;
	protected gameLink: string;

	constructor(
		protected game: GameStateService,
		private apiService: ApiService
	) {
		this.gameLink = environment.gameLink
	}

	openJoinLink(): void {
		window.open('join/' + this.game.Name(), '_blank')
	}

	copyJoinLink(): void {
		let joinLink = `${this.gameLink}/join/${this.game.Name()}`
		navigator.clipboard.writeText(joinLink).then(function () { }, function (err) { })
	}

	addBot(): void {
		this.apiService.AddBot().subscribe()
	}
}
