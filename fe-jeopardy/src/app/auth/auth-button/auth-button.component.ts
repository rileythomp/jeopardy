import { Component, Input } from '@angular/core';

@Component({
	selector: 'app-auth-button',
	templateUrl: './auth-button.component.html',
	styleUrls: ['./auth-button.component.less']
})
export class AuthButtonComponent {
	@Input() buttonText: string

	constructor() { }
}
