import { Component, Input } from '@angular/core';
import { Observer } from 'rxjs';
import { environment } from '../../../environments/environment';
import { ServerUnavailableMsg } from '../../model/model';
import { ApiService } from '../../services/api.service';
import { GameStateService } from '../../services/game-state.service';
import { ModalService } from '../../services/modal.service';

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
		private api: ApiService,
		private modal: ModalService
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
		this.api.AddBot().subscribe(this.handleResp())
	}

	startGame(): void {
		this.api.StartGame().subscribe(this.handleResp())
	}

	private handleResp(): Partial<Observer<any>> {
		return {
			next: () => { },
			error: (err: any) => {
				let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg;
				this.modal.displayMessage(msg)
			}
		}
	}
}
