import { Component, Input } from '@angular/core';
import { Player } from 'src/app/model/model';
import { GameStateService } from 'src/app/services/game-state.service';

@Component({
	selector: 'app-player-podium',
	templateUrl: './player-podium.component.html',
	styleUrls: ['./player-podium.component.less']
})
export class PlayerPodiumComponent {
	@Input() player: Player; 
	@Input() scoreChanges: any = {}

	constructor(protected game: GameStateService) { }

	abs(num: number): number {
		if (!num) {
			return 0
		}
		return Math.abs(num)
	}

	scoreChange(): number {
		return this.scoreChanges[this.player.id] ?? 0
	}

	onClock(player: Player): boolean {
		return player.canPick || player.canAnswer || player.canWager
	}
}
