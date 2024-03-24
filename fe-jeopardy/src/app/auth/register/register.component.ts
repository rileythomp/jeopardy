import { Component, ElementRef, ViewChild } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';
import { ModalService } from 'src/app/services/modal.service';
import { StorageService } from 'src/app/services/storage.service';
import { NewPasswordComponent } from '../new-password/new-password.component';

@Component({
	selector: 'app-register',
	templateUrl: './register.component.html',
	styleUrl: './register.component.less'
})
export class RegisterComponent {
	protected uploadedImg: boolean = false;
	protected imgUpload: boolean = false;
	protected uploadedImgUrl: string = '';
	protected email: string = '';
	protected username: string = '';
	protected registerDone: boolean = false
	@ViewChild('userImg') userImg: ElementRef
	@ViewChild('usernameInput') usernameInput: ElementRef
	@ViewChild('emailInput') emailInput: ElementRef
	@ViewChild(NewPasswordComponent) passwords: NewPasswordComponent

	constructor(
		private storage: StorageService,
		private auth: AuthService,
		protected modal: ModalService,
	) { }

	showImgUpload() {
		this.imgUpload = !this.imgUpload
	}

	async handleUserUpload(event: any) {
		let file = event.target.files[0]
		let { url, error } = await this.storage.UploadImg('anon', file)
		if (error) {
			this.userImg.nativeElement.value = ''
			this.userImg.nativeElement.style.border = '1px solid red'
			setTimeout(() => {
				this.userImg.nativeElement.style.border = '1px solid grey'
			}, 2000)
			return
		}
		this.uploadedImg = true
		this.uploadedImgUrl = url
	}

	private validUsername(): boolean {
		return this.username.length > 0 && this.username.length < 50
	}

	private validEmail(): boolean {
		return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(this.email)
	}

	protected usernameBorder(border: string) {
		this.usernameInput.nativeElement.style.border = border
	}

	protected emailBorder(border: string) {
		this.emailInput.nativeElement.style.border = border
	}

	async createAccount() {
		let hasError = false
		if (!this.validUsername()) {
			this.usernameBorder('1px solid red')
			hasError = true
		}
		if (!this.validEmail()) {
			this.emailBorder('1px solid red')
			hasError = true
		}
		if (!this.passwords.isValid()) {
			this.passwords.setBorder('1px solid red')
			hasError = true
		}
		if (hasError) {
			return
		}
		let error = await this.auth.SignUp(this.email, this.passwords.password, this.username, this.uploadedImgUrl)
		if (error) {
			console.error(error)
			this.modal.displayMessage('Uh oh, we\'re not able to create new accounts right now. Please try again later.')
			return
		}
		this.usernameInput.nativeElement.disabled = true
		this.emailInput.nativeElement.disabled = true
		this.passwords.disableInput()
		this.registerDone = true
	}
}
