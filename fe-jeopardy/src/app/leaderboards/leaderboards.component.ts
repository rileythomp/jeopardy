import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { ApiService } from '../services/api.service';
import { ModalService } from '../services/modal.service';

@Component({
	selector: 'app-leaderboards',
	templateUrl: './leaderboards.component.html',
	styleUrl: './leaderboards.component.less'
})
export class LeaderboardsComponent {
	protected leaderboard: any[]
	protected leaderboardType: string = 'correct_rate'

	constructor(
		private api: ApiService,
		private modal: ModalService,
		protected router: Router
	) {

	}

	async ngOnInit() {
		await this.updateLeaderboard()
	}

	async updateLeaderboard() {
		let { leaderboard, err } = await this.api.GetLeaderboard(this.leaderboardType)
		if (err) {
			this.modal.displayMessage('Uh oh, we were unable to get the leaderboards right now. Please try again later.')
			return
		}
		this.leaderboard = leaderboard
	}
}
