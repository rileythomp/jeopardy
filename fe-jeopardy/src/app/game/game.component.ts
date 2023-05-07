import { Component, OnInit } from '@angular/core';
import { GameStateService } from '../game-state.service';
import { WebsocketService } from '../websocket.service';
import { JwtService } from '../jwt.service';

@Component({
	selector: 'app-game',
	templateUrl: './game.component.html',
	styleUrls: ['./game.component.less'],
})
export class GameComponent implements OnInit {
	playerNames: string[];

	constructor(
		private gameStateService: GameStateService,
		private websocketService: WebsocketService,
		private jwtService: JwtService,
	) { }

	ngOnInit(): void {
		this.playerNames = this.gameStateService.playerNames();
		// this.websocketService.send({"hello": "world"})
		// this.websocketService.onmessage((event: { data: string; }) => {
		// 	console.log("received game message")
		// 	console.log(JSON.parse(event.data))
		// })

	}

}
