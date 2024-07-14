import { Component, Input } from '@angular/core';

@Component({
	selector: 'app-pw-req',
	templateUrl: './pw-req.component.html',
	styleUrl: './pw-req.component.less'
})
export class PwReqComponent {
	@Input() requirement: string;
	@Input() satisfied: boolean;
}
