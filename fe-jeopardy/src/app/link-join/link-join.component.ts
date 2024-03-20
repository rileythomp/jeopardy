import { Component } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ServerUnavailableMsg } from '../model/model';
import { ApiService } from '../services/api.service';
import { AuthService } from '../services/auth.service';
import { JwtService } from '../services/jwt.service';
import { ModalService } from '../services/modal.service';

@Component({
    selector: 'app-link-join',
    templateUrl: './link-join.component.html',
    styleUrls: ['./link-join.component.less']
})
export class LinkJoinComponent {
    protected gameCode: string;
    protected playerName: string;

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private apiService: ApiService,
        private jwtService: JwtService,
        private modal: ModalService,
        private user: AuthService,
    ) {
        this.gameCode = this.route.snapshot.paramMap.get('gameCode') ?? '';
    }

    joinGame() {
        let playerImg = ''
        if (this.user.Authenticated()) {
            this.playerName = this.user.Name()
            playerImg = this.user.ImgUrl()
        }
        this.apiService.JoinGameByCode(this.playerName, playerImg, this.gameCode).subscribe({
            next: (resp: any) => {
                this.jwtService.SetJWT(resp.token);
                this.router.navigate([`/game/${resp.game.name}`]);
            },
            error: (err: any) => {
                let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg;
                this.modal.displayMessage(msg)
            }
        });
    }
}