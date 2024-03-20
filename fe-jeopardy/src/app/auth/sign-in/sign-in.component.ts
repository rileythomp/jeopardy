import { Component, EventEmitter, Output } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';

@Component({
	selector: 'app-sign-in',
	templateUrl: './sign-in.component.html',
	styleUrls: ['./sign-in.component.less']
})
export class SignInComponent {
	@Output() signInError = new EventEmitter<boolean>();

	constructor(private auth: AuthService) { }

	async signIn(provider: string) {
		if (await this.auth.SignIn(provider)) {
			this.signInError.emit(true)
		}
	}
}
