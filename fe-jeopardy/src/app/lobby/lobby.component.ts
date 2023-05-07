import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { WebsocketService } from '../websocket.service';
import { JwtService } from '../jwt.service';
import { GameStateService } from '../game-state.service';

@Component({
	selector: 'app-lobby',
	templateUrl: './lobby.component.html',
	styleUrls: ['./lobby.component.less']
})
export class LobbyComponent implements OnInit {
	lobbyMessage: string;
	jwt: string;
	playerNames: string[];

	constructor(
		private router: Router,
		private websocketService: WebsocketService,
		private jwtService: JwtService,
		private gameStateService: GameStateService,
	) { }

	ngOnInit(): void {
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt;
		});

		this.websocketService.connect('ws://localhost:8080/jeopardy/play');

		this.websocketService.onopen(() => {
			let playReq = {
				token: this.jwt,
			}
			this.websocketService.send(playReq);
		})

		this.websocketService.onmessage((event: { data: string; }) => {
			let resp = JSON.parse(event.data);
			if (resp.code == 200) {
				this.lobbyMessage = resp.message;
				this.gameStateService.updateGameState(resp.game);
				this.playerNames = this.gameStateService.playerNames();
				if (this.gameStateService.readyToPlay()) {
					this.router.navigate(['/game']);
				}
			} else {
				alert('Unable to join game');
			}
		})
	}

}
