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
    protected joinCode: string
    protected playerName: string
    private playerImg: string = ''
    private playerEmail: string = ''

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private api: ApiService,
        private jwt: JwtService,
        private modal: ModalService,
        private auth: AuthService,
    ) {
        this.joinCode = this.route.snapshot.paramMap.get('joinCode') ?? '';
    }

    ngOnInit() {
        this.auth.user.subscribe(user => {
            this.playerImg = user.imgUrl
            this.playerName = user.name
            this.playerEmail = user.email
            this.joinGame()
        })
    }

    joinGame() {
        this.api.JoinGameByCode(this.playerName, this.playerImg, this.playerEmail, this.joinCode).subscribe({
            next: (resp: any) => {
                this.jwt.SetJWT(resp.token);
                this.router.navigate([`/game/${resp.game.name}`]);
            },
            error: (err: any) => {
                let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg;
                this.modal.displayMessage(msg)
            }
        });
    }
}