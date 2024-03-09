import { Component } from '@angular/core';
import { GameStateService } from '../services/game-state.service';
import { WebsocketService } from '../services/websocket.service';
import { PlayerService } from '../services/player.service';

@Component({
    selector: 'app-modal',
    templateUrl: './modal.component.html',
    styleUrls: ['./modal.component.less']
})
export class ModalComponent {
    protected showInfo: boolean = false
    protected showModal: boolean = false
    protected message: string
    private modalTimeout: NodeJS.Timeout
    protected firstTime: boolean

    constructor(
        protected game: GameStateService,
        protected player: PlayerService,
        private websocket: WebsocketService
    ) { }

    showMessage(msg: string) {
        clearTimeout(this.modalTimeout)
        this.message = msg
        this.showModal = true
        this.modalTimeout = setTimeout(() => {
            this.showModal = false
        }, 10000)
    }

    showJeopardyInfo(firstTime: boolean, show: boolean) {
        this.showInfo = show
        this.firstTime = firstTime
    }

    disputeQuestion(dispute: boolean) {
        this.websocket.Send({
            state: this.game.State(),
            dispute: dispute,
        })
    }
}
