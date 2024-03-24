import { Component } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';
import { StorageService } from 'src/app/services/storage.service';
import { User } from '../model/model';
import { ModalService } from '../services/modal.service';

@Component({
	selector: 'app-profile',
	templateUrl: './profile.component.html',
	styleUrl: './profile.component.less'
})
export class ProfileComponent {
	protected user: User
	protected showImgUpload: boolean = false
	protected showPasswordReset: boolean = false

	constructor(
		private auth: AuthService,
		private storage: StorageService,
		private modal: ModalService,
	) {
		this.auth.user.subscribe(user => {
			this.user = user
		})
		this.auth.GetUser();
	}

	async changeProfilePicture(event: any) {
		let { url, error } = await this.storage.UploadImg(event.target.files[0])
		if (error) {
			this.modal.displayMessage('Uh oh, there was an error upadting your profile picture. Please try again later.')
			return
		}
		error = await this.auth.UpdateUserImg(url)
		if (error) {
			this.modal.displayMessage('Uh oh, there was an error upadting your profile picture. Please try again later.')
			return
		}
		this.showImgUpload = false
		this.auth.GetUser()
	}

	async sendPasswordResetEmail() {
		let error = await this.auth.SendPasswordResetEmail(this.user.email)
		if (error) {
			this.modal.displayMessage('Uh oh, there was an error restting your password. Please try again later.')
			return
		}
		this.modal.displayMessage('Password reset email sent. Please check your email.')
	}
}
