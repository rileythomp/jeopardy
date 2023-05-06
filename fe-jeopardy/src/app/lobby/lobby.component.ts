import { Component, OnInit } from '@angular/core';
import { WebsocketService } from '../websocket.service';
import { JwtService } from '../jwt.service';

@Component({
	selector: 'app-lobby',
	templateUrl: './lobby.component.html',
	styleUrls: ['./lobby.component.less']
})
export class LobbyComponent implements OnInit {
	lobbyMessage: string;
	jwt: string;

	constructor(
		private websocketService: WebsocketService,
		private jwtService: JwtService,
	) { 
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt;
		});
	}

	ngOnInit(): void {
		this.websocketService.connect('ws://localhost:8080/jeopardy/play');

		this.websocketService.onopen(() => {
			let playReq = {
				token: this.jwt,
			}
			this.websocketService.send(playReq);
		})

		this.websocketService.onmessage((event: { data: string; }) => {
			let response = JSON.parse(event.data);
			if (response.code == 200) {
				this.lobbyMessage = response.message;
			} else {
				alert('Unable to join game');
			}
		})
	}

}
