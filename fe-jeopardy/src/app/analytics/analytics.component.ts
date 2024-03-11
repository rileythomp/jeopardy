import { Component, OnInit } from '@angular/core';
import { ApiService } from '../services/api.service';
import { ModalService } from '../services/modal.service';

@Component({
	selector: 'app-analytics',
	templateUrl: './analytics.component.html',
	styleUrls: ['./analytics.component.less']
})
export class AnalyticsComponent implements OnInit {
	gamesPlayed: number
	firstRoundScore: number
	secondRoundScore: number
	firstRoundAnsRate: number
	firstRoundCorrRate: number
	secondRoundAnsRate: number
	secondRoundCorrRate: number

	constructor(
		private apiService: ApiService,
		protected modal: ModalService,
	) { }

	ngOnInit() {
		this.apiService.GetAnalytics().subscribe((resp: any) => {
			this.gamesPlayed = resp.gamesPlayed
			this.firstRoundScore = resp.firstRoundScore
			this.secondRoundScore = resp.secondRoundScore
			this.firstRoundAnsRate = resp.firstRoundAnsRate
			this.firstRoundCorrRate = resp.firstRoundCorrRate
			this.secondRoundAnsRate = resp.secondRoundAnsRate
			this.secondRoundCorrRate = resp.secondRoundCorrRate
		})
	}
}
