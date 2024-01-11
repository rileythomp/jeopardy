import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { JwtService } from '../jwt.service';
import { ApiService } from '../api.service';
import { GameState as GameState } from '../model/model';
// import { generateFakeWordByLength } from 'fakelish';

@Component({
	selector: 'app-join',
	templateUrl: './join.component.html',
	styleUrls: ['./join.component.less'],
})
export class JoinComponent implements OnInit {
	title: string = 'Jeopardy';
	playerName: string = '';
	gameName: string = '';
	jwt: string;

	constructor(
		private router: Router,
		private jwtService: JwtService,
		private apiService: ApiService, 
	) { }

	ngOnInit(): void {
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt;
		});

		// (async ()=>{
		// 	this.playerName = await generateFakeWordByLength(7);
		// 	this.gameName = 'autojoined'
		// 	this.joinGame(false);
		// })();
	}

	joinGame(privateGame: boolean) {
		this.apiService.joinGame(this.playerName, this.gameName, privateGame).subscribe({
			next: (resp: any) => {
				this.jwtService.setJwt(resp.token); 
				if (resp.game.state in GameState) {
					this.router.navigate(['/lobby']);
				} else {
					alert('Unable to join the game');
				}
			},
			error: (err: any) => {
				alert('Unable to join the game');
			},
		});
	}

	rejoin() {
		this.router.navigate(['/lobby']);
	}
}
