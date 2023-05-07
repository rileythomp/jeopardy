import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { JwtService } from '../jwt.service';
import { WebsocketService } from '../websocket.service';
import { GameState as GameState } from '../model/model';

@Component({
	selector: 'app-join',
	templateUrl: './join.component.html',
	styleUrls: ['./join.component.less'],
	providers: [WebsocketService],
})
export class JoinComponent implements OnInit {
	title: string = 'Jeopardy';
	playerName: string = '';
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
		this.websocketService.connect('ws://localhost:8080/jeopardy/join')
	}

	joinGame() {
		let joinReq = {
			playerName: this.playerName,
		}
		this.websocketService.send(joinReq);

		this.websocketService.onmessage((event: { data: string; }) => {
			let resp = JSON.parse(event.data);
			this.jwtService.setJwt(resp.token);
			if (resp.game.state == GameState.PreGame) {
				this.router.navigate(['/lobby']);
			} else {
				alert('Unable to join game');
			}
		})
	}

}
