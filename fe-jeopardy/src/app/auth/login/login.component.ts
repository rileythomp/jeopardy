import { Component, ElementRef, ViewChild } from '@angular/core';
import { SignInWithPasswordCredentials } from '@supabase/supabase-js';
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
	protected invalidLogin: boolean = false
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
		if (!this.email) {
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
		let credentials: SignInWithPasswordCredentials = {
			email: this.email,
			password: this.password,
		}
		let error = await this.auth.SignInWithPassword(credentials)
		if (error) {
			this.invalidLogin = true
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
}
