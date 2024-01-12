import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { JwtService } from '../services/jwt.service';
import { ApiService } from '../services/api.service';
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
		// 	this.joinGame(true);
		// })();
	}

	joinGame(privateGame: boolean) {
		this.apiService.JoinGame(this.playerName, this.gameName, privateGame).subscribe({
			next: (resp: any) => {
				this.jwtService.SetJWT(resp.token); 
				this.router.navigate(['/game']);
			},
			error: (resp: any) => {
				// TODO: REPLACE WITH MODAL
				alert(resp.error.message);
			},
		});
	}

	rejoin() {
		this.router.navigate(['/game']);
	}
}
