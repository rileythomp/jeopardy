import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { JwtService } from '../jwt.service';
import { WebsocketService } from '../websocket.service';

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
	) {
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt;
		});
	 }

	ngOnInit(): void {
		this.websocketService.connect('ws://localhost:8080/jeopardy/join')
	}

	joinGame() {
		let joinReq = {
			playerName: this.playerName,
		}
		this.websocketService.send(joinReq);

		this.websocketService.onmessage((event: { data: string; }) => {
			let response = JSON.parse(event.data);
			this.jwtService.setJwt(response.token);
			if (response.code == 200) {
				this.router.navigate(['/lobby']);
			} else {
				alert('Unable to join game');
			}
		})
	}

}
