import { Component, OnInit } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';
import { StorageService } from 'src/app/services/storage.service';
import { User } from '../model/model';
import { ApiService } from '../services/api.service';
import { ModalService } from '../services/modal.service';

@Component({
	selector: 'app-profile',
	templateUrl: './profile.component.html',
	styleUrl: './profile.component.less'
})
export class ProfileComponent implements OnInit {
	protected user: User
	protected showImgUpload: boolean = false
	protected showPasswordReset: boolean = false
	protected editName = false
	protected editedUserName = ''
	protected playSomeGames = false
	protected analytics: {
		wins: number,
		games: number,
		points: number,
		answers: number,
		correct: number,
		maxPoints: number,
		maxCorrect: number,
		winPercentage: number,
		correctPercentage: number,
	}

	constructor(
		private auth: AuthService,
		private storage: StorageService,
		private modal: ModalService,
		private api: ApiService,
	) {
		this.auth.user.subscribe(user => {
			this.user = user
			this.editedUserName = this.user.name
		})
	}

	async ngOnInit() {
		await this.auth.GetUser()
		this.api.GetPlayerAnalytics(this.user.email).subscribe({
			next: (resp: any) => {
				this.analytics = resp
				this.analytics.winPercentage = Math.round((resp.wins / resp.games) * 1000) / 10
				this.analytics.correctPercentage = Math.round((resp.correct / resp.answers) * 1000) / 10
			},
			error: (err: any) => {
				console.error(err)
				this.playSomeGames = true
				this.analytics = {
					wins: 0,
					games: 0,
					points: 0,
					answers: 0,
					correct: 0,
					maxPoints: 0,
					maxCorrect: 0,
					winPercentage: 0,
					correctPercentage: 0,
				}
			}
		})
	}

	async changeProfilePicture(event: any) {
		let { url, error } = await this.storage.UploadImg(this.user.email, event.target.files[0])
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

	showEditName(edit: boolean) {
		if (!edit) {
			this.editedUserName = this.user.name
		}
		this.editName = edit
	}

	async updateUserName() {
		if (await this.auth.UpdateUserName(this.editedUserName)) {
			this.modal.displayMessage('Uh oh, there was an error updating your name. Please try again later.')
			return
		}
		this.user.name = this.editedUserName
		this.showEditName(false)
	}
}
