import { Component } from '@angular/core';
import { ModalService } from '../services/modal.service';

@Component({
	selector: 'app-game-message',
	templateUrl: './game-message.component.html',
	styleUrls: ['./game-message.component.less']
})
export class GameMessageComponent {
	constructor(protected modal: ModalService) { }
}
