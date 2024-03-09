import { Component, OnInit } from '@angular/core';
import { ApiService } from '../services/api.service';

@Component({
	selector: 'app-analytics',
	templateUrl: './analytics.component.html',
	styleUrls: ['./analytics.component.less']
})
export class AnalyticsComponent implements OnInit {
	showAnalytics: boolean
	gamesPlayed: number
	firstRoundAns: number
	firstRoundCorr: number
	secondRoundAns: number
	secondRoundCorr: number

	constructor(
		private apiService: ApiService
	) { }

	ngOnInit(): void {
		this.gamesPlayed = 3
	}

	toggleAnalytics(show: boolean) {
		this.showAnalytics = show
	}
}
