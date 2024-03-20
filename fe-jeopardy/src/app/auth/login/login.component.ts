import { Component, EventEmitter, Output } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';

@Component({
	selector: 'app-login',
	templateUrl: './login.component.html',
	styleUrls: ['./login.component.less']
})
export class LoginComponent {
	@Output() signInError = new EventEmitter<boolean>();

	constructor(private auth: AuthService) { }

	async googleSignIn() {
		if (await this.auth.SignIn()) {
			this.signInError.emit(true)
		}
	}
}
