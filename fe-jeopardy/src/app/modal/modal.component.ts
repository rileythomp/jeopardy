import { Component } from '@angular/core';
import { GameStateService } from '../services/game-state.service';
import { WebsocketService } from '../services/websocket.service';

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
    public showDisputeModal: boolean = false
    protected votedDispute: boolean = false

    constructor(
        protected game: GameStateService,
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

    showJeopardyInfo(firstTime: boolean) {
        this.showInfo = true
        this.firstTime = firstTime
    }

    disputeQuestion(dispute: boolean) {
        this.websocket.Send({
            state: this.game.State(),
            dispute: dispute,
        })
        this.votedDispute = true
    }

    showDispute() {
        this.showDisputeModal = true
        this.votedDispute = false
    }

    hideDispute() {
        this.showDisputeModal = false
        this.votedDispute = false
    }
}
