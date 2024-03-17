import { Component, Input, OnInit } from '@angular/core';
import { ServerUnavailableMsg } from 'src/app/model/model';
import { ApiService } from 'src/app/services/api.service';
import { ModalService } from 'src/app/services/modal.service';
import { JwtService } from 'src/app/services/jwt.service';
import { Router } from '@angular/router';

@Component({
	selector: 'app-config',
	templateUrl: './config.component.html',
	styleUrls: ['./config.component.less']
})
export class ConfigComponent {
	@Input() playerName: string
	protected oneRoundChecked: boolean = true
	protected twoRoundChecked: boolean = false
	protected penaltyChecked: boolean = true
	protected botConfig: number = 0
	protected pickConfig: number = 30
	protected buzzConfig: number = 30
	protected answerConfig: number = 15
	protected voteConfig: number = 10
	protected wagerConfig: number = 30
	protected firstRoundCategories: any[] = []
	protected secondRoundCategories: any[] = []
	protected searchResults: any[] = []
	protected categorySearch: string = ''
	protected searchLoader = false
	protected questionMode: string = 'cyo'

	constructor(
		private apiService: ApiService,
		private modal: ModalService,
		private jwtService: JwtService,
		private router: Router
	) { }

	searchCategories() {
		this.searchLoader = true
		if (this.categorySearch == '') {
			this.searchResults = []
			document.getElementById('results-dropdown')!.style.borderBottom = 'none';
			this.searchLoader = false
			return
		}
		this.apiService.SearchCategories(this.categorySearch, this.twoRoundChecked).subscribe({
			next: (resp: any) => {
				this.searchResults = resp
				if (resp.length > 0) {
					document.getElementById('results-dropdown')!.style.borderBottom = '1px solid black';
				} else {
					document.getElementById('results-dropdown')!.style.borderBottom = 'none';
				}
				this.searchLoader = false
			},
			error: (err: any) => {
				let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg;
				this.searchResults = [msg]
			}
		})
	}

	canAddCategory(category: any, round: number, list: any[]): boolean {
		return category.round == round && list.length < 6 && !list.some((c) => c.category == category.category && c.airDate == category.airDate)
	}

	addCategory(category: any) {
		if (this.canAddCategory(category, 1, this.firstRoundCategories)) {
			this.firstRoundCategories.push(category)
		} else if (this.canAddCategory(category, 2, this.secondRoundCategories)) {
			this.secondRoundCategories.push(category)
		}
		this.searchResults = []
		document.getElementById('results-dropdown')!.style.borderBottom = 'none';
	}

	removeCategory(category: any) {
		this.firstRoundCategories = this.firstRoundCategories.filter((c) => c.category != category.category && c.airDate != category.airDate)
		this.secondRoundCategories = this.secondRoundCategories.filter((c) => c.category != category.category && c.airDate != category.airDate)
	}

	validateBotConfig() {
		this.botConfig = Math.min(Math.max(this.botConfig, 0), 2)
	}

	validatePickConfig() {
		this.pickConfig = Math.min(Math.max(this.pickConfig, 3), 60)
	}

	validateBuzzConfig() {
		this.buzzConfig = Math.min(Math.max(this.buzzConfig, 10), 60)
	}

	validateAnswerConfig() {
		this.answerConfig = Math.min(Math.max(this.answerConfig, 3), 60)
	}

	validateVoteConfig() {
		this.voteConfig = Math.min(Math.max(this.voteConfig, 3), 60)
	}

	validateWagerConfig() {
		this.wagerConfig = Math.min(Math.max(this.wagerConfig, 3), 60)
	}

	hideConfig() {
		this.modal.hideConfig()
	}

	createPrivateGame(bots: number) {
		this.apiService.CreatePrivateGame(
			this.playerName,
			bots, this.twoRoundChecked, this.penaltyChecked,
			this.pickConfig, this.buzzConfig, this.answerConfig, this.voteConfig, this.wagerConfig,
			this.firstRoundCategories, this.secondRoundCategories
		).subscribe({
			next: (resp: any) => {
				this.jwtService.SetJWT(resp.token)
				this.router.navigate([`/game/${resp.game.name}`])
			},
			error: (err: any) => {
				let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg;
				this.modal.displayMessage(msg)
			}
		})
	}
}
