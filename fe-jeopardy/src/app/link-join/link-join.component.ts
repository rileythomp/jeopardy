import { Component, ViewChild } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { ApiService } from '../services/api.service';
import { JwtService } from '../services/jwt.service';
import { ServerUnavailableMsg } from '../constants';
import { ModalComponent } from '../modal/modal.component';

@Component({
    selector: 'app-link-join',
    templateUrl: './link-join.component.html',
    styleUrls: ['./link-join.component.less']
})
export class LinkJoinComponent {
    protected gameCode: string;
    protected playerName: string;

    @ViewChild(ModalComponent) private modal: ModalComponent

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
                let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg;
                this.modal.showMessage(msg)
            }
        });
    }
}