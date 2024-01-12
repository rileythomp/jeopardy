import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { WebsocketService } from '../websocket.service';
import { JwtService } from '../jwt.service';
import { GameStateService } from '../game-state.service';
import { PlayerService } from '../player.service';
import { Player, GameState, Ping } from '../model/model';

@Component({
	selector: 'app-lobby',
	templateUrl: './lobby.component.html',
	styleUrls: ['./lobby.component.less']
})
export class LobbyComponent implements OnInit {
	lobbyMessage: string;
	jwt: string;
	players: Player[];
	playerName: string;
	gameName: string;

	constructor(
		private router: Router,
		private websocketService: WebsocketService,
		private jwtService: JwtService,
		private gameState: GameStateService,
		private player: PlayerService,
	) { }

	ngOnInit(): void {
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt;
		});

		this.websocketService.connect('play');

		this.websocketService.onopen(() => {
			let playReq = {
				token: this.jwt,
			}
			this.websocketService.send(playReq);
		})

		this.websocketService.onmessage((event: { data: string; }) => {
			let resp = JSON.parse(event.data);

			if (resp.code != 200) {
				alert(resp.message);
				this.router.navigate(['/join']);
				return
			}

			if (resp.message == Ping) {
				return
			}

			if (!(resp.game.state in GameState)) {
				alert('Unable to start the game');
				return
			}

			this.lobbyMessage = resp.message;
			this.gameState.updateGameState(resp.game);
			this.players = this.gameState.getPlayers();
			this.player.updatePlayer(resp.curPlayer);

			if (resp.game.state == GameState.PreGame) {
				this.playerName = this.player.getName();
				this.gameName = this.gameState.getName();
			} else {
				this.router.navigate(['/game']);
			}
		})
	}
}
