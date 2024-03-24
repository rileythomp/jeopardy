import { Component } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ServerUnavailableMsg, User } from '../model/model';
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
    protected user: User

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
            this.user = user
            if (this.user.authenticated) {
                this.joinGame()
            }
        })
    }

    joinGame() {
        this.api.JoinGameByCode(this.user.name, this.user.imgUrl, this.user.email, this.joinCode).subscribe({
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