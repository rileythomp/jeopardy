import { Component, ElementRef, ViewChild } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';
import { ModalService } from 'src/app/services/modal.service';

@Component({
	selector: 'app-login',
	templateUrl: './login.component.html',
	styleUrl: './login.component.less'
})
export class LoginComponent {
	protected password: string = '';
	protected email: string = '';
	protected showEye: boolean = true
	protected loginMessage: string = ''
	protected showSendReset: boolean = false
	@ViewChild('emailInput') emailInput: ElementRef
	@ViewChild('passwordInput') passwordInput: ElementRef

	constructor(
		private auth: AuthService,
		protected modal: ModalService,
	) { }

	protected emailBorder(border: string) {
		this.emailInput.nativeElement.style.border = border
	}

	protected passwordsBorder(border: string) {
		this.passwordInput.nativeElement.style.border = border
	}

	async logIn() {
		let hasError = false
		if (!this.validEmail()) {
			this.emailBorder('1px solid red')
			hasError = true
		}
		if (!this.password) {
			this.passwordsBorder('1px solid red')
			hasError = true
		}
		if (hasError) {
			return
		}
		if (await this.auth.SignInWithPassword(this.email, this.password)) {
			this.loginMessage = 'Sorry, the email or password you entered is incorrect. Please try again.'
			this.emailBorder('1px solid red')
			this.passwordsBorder('1px solid red')
			return
		}
		location.replace('');
	}

	protected showPassword(show: boolean) {
		this.passwordInput.nativeElement.type = show ? 'text' : 'password'
		this.showEye = !show
	}

	protected async sendPasswordResetEmail() {
		if (!this.validEmail()) {
			this.emailBorder('1px solid red')
			return
		}
		if (await this.auth.SendPasswordResetEmail(this.email)) {
			this.loginMessage = 'Sorry, there was an error sending the password reset email. Please try again later.'
			return
		}
		this.loginMessage = 'Password reset email sent.'
		this.showSendReset = false
	}

	private validEmail(): boolean {
		return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(this.email)
	}

}
