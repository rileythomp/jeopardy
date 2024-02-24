import { Component } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { ApiService } from '../services/api.service';
import { JwtService } from '../services/jwt.service';
import { ServerUnavailableMessage } from '../constants';

@Component({
    selector: 'app-link-join',
    templateUrl: './link-join.component.html',
    styleUrls: ['./link-join.component.less']
})
export class LinkJoinComponent {
    gameCode: string;
    playerName: string;

    protected showModal: boolean = false;
    protected modalMessage: string;
    private modalTimeout: NodeJS.Timeout

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
            error: (err: any) => {
                this.showMessage(err)
            }
        });
    }

    showMessage(err: any) {
        clearTimeout(this.modalTimeout)
        this.modalMessage = err.status != 0 ? err.error.message : ServerUnavailableMessage;
        this.showModal = true;
        this.modalTimeout = setTimeout(() => {
            this.showModal = false
        }, 10000)
    }
}