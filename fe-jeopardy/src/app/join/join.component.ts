import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { JwtService } from '../jwt.service';
import { WebsocketService } from '../websocket.service';
import { GameState as GameState } from '../model/model';
import { environment } from 'src/environments/environment';

@Component({
	selector: 'app-join',
	templateUrl: './join.component.html',
	styleUrls: ['./join.component.less'],
	providers: [WebsocketService],
})
export class JoinComponent implements OnInit {
	title: string = 'Jeopardy';
	playerName: string = '';
	gameName: string = '';
	jwt: string;

	constructor(
		private router: Router,
		private websocketService: WebsocketService,
		private jwtService: JwtService,
	) { }

	ngOnInit(): void {
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt;
		});
		this.websocketService.connect(`${environment.websocketProtocol}://${environment.apiServerUrl}/jeopardy/join`)
		// TODO: REMOVE THIS
		this.websocketService.onopen(() => {
			this.playerName = this.generateRandomString(7);
			this.gameName = 'testroom'
			this.joinGame();
		})
	}

	joinGame() {
		let joinReq = {
			playerName: this.playerName,
			gameName: this.gameName,
		}
		this.websocketService.send(joinReq);

		this.websocketService.onmessage((event: { data: string; }) => {
			let resp = JSON.parse(event.data);
			this.jwtService.setJwt(resp.token);
			if (resp.game.state in GameState) {
				this.router.navigate(['/lobby']);
			} else {
				alert('Unable to join the lobby');
			}
		})
	}

	generateRandomString(length: number): string {
		const characters = 'abcdefghijklmnopqrstuvwxyz0123456789';
		let result = '';
		for (let i = 0; i < length; i++) {
		  const randomIndex = Math.floor(Math.random() * characters.length);
		  result += characters.charAt(randomIndex);
		}
		return result;
	  }

}
