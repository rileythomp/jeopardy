import { Component } from '@angular/core';
import { GameStateService } from '../../services/game-state.service';
import { Player } from '../../model/model';
import { PlayerService } from '../../services/player.service';
import { WebsocketService } from '../../services/websocket.service';
import { ApiService } from '../../services/api.service';

@Component({
    selector: 'app-post-game',
    templateUrl: './post-game.component.html',
    styleUrls: ['./post-game.component.less']
})
export class PostGameComponent {

    constructor(
        private apiService: ApiService,
        private websocketService: WebsocketService,
        protected game: GameStateService,
        protected player: PlayerService,
    ) { }

    canProtestForPlayer(player: Player): boolean {
        return !Object.keys(player.finalProtestors).includes(this.player.Id());
    }

    protestFinalCorrectness(playerId: string) {
		this.websocketService.Send({protestFor: playerId});
	}

    playAgain() {
		return this.apiService.PlayAgain({ "hello": "world" }).subscribe({
			next: (resp: any) => {
				console.log('playing again', resp)
			},
			error: (err: any) => {
				console.log('Error playing again', err)
				alert('Error playing again')
			},
		})
	}

	leaveGame() {
		return this.apiService.LeaveGame({ "hello": "world" }).subscribe({
			next: (resp: any) => {
				console.log('left game', resp)
			},
			error: (err: any) => {
				console.log('Error leaving game', err)
				alert('Error leaving game')
			},
		})
	}
}
