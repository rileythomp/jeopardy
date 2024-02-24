import { Component } from '@angular/core';

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

    showMessage(msg: string) {
        clearTimeout(this.modalTimeout)
        this.message = msg
        this.showModal = true
        this.modalTimeout = setTimeout(() => {
            this.showModal = false
        }, 10000)
    }

    showJeopardyInfo() {
        this.showInfo = true
    }
}
