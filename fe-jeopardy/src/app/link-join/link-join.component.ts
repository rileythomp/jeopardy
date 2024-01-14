import { Component } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { ApiService } from '../services/api.service';
import { JwtService } from '../services/jwt.service';

@Component({
    selector: 'app-link-join',
    templateUrl: './link-join.component.html',
    styleUrls: ['./link-join.component.less']
})
export class LinkJoinComponent {
    gameCode: string;
    playerName: string;

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private apiService: ApiService,
        private jwtService: JwtService,
    ) {
        this.gameCode = this.route.snapshot.paramMap.get('gameCode') ?? '';
    }

    joinGame(playerName: string, gameCode: string) {
        this.apiService.JoinGameByCode(playerName, gameCode).subscribe({
            next: (resp: any) => {
                this.jwtService.SetJWT(resp.token);
                this.router.navigate([`/game/${resp.game.name}`]);
            },
            error: (resp: any) => {
                // TODO: REPLACE WITH MODAL
                alert(resp.error.message);
                this.router.navigate(['/join'])
            }
        });
    }
}