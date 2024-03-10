import { Component } from '@angular/core';
import { ModalService } from '../services/modal.service';

@Component({
	selector: 'app-instructions',
	templateUrl: './instructions.component.html',
	styleUrls: ['./instructions.component.less']
})
export class InstructionsComponent {
	constructor(protected modal: ModalService) { }
}
