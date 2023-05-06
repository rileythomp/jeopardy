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
		console.log('jwt in lobby component: ' + this.jwt);
		this.websocketService.connect('ws://localhost:8080/jeopardy/play');

		this.websocketService.ws.onopen = () => {
			console.log('websocket connection opened');
			let playReq = {
				token: this.jwt,
			}
			this.websocketService.ws.send(JSON.stringify(playReq));
		};

		this.websocketService.ws.onmessage = (event: { data: string; }) => {
			let response = JSON.parse(event.data);
			console.log(response);
			if (response.code == 200) {
				this.lobbyMessage = response.message;
			} else {
				alert('Unable to join game');
			}
		};
	}

}
