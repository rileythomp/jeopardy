import { Component, Input } from '@angular/core';
import { Router } from '@angular/router';
import { ServerUnavailableMsg } from 'src/app/model/model';
import { ApiService } from 'src/app/services/api.service';
import { AuthService } from 'src/app/services/auth.service';
import { JwtService } from 'src/app/services/jwt.service';
import { ModalService } from 'src/app/services/modal.service';

@Component({
	selector: 'app-config',
	templateUrl: './config.component.html',
	styleUrls: ['./config.component.less']
})
export class ConfigComponent {
	@Input() playerName: string
	@Input() oneRoundChecked: boolean = true
	@Input() twoRoundChecked: boolean = false
	@Input() penaltyChecked: boolean = true
	protected botConfig: number = 0
	protected pickConfig: number = 30
	protected buzzConfig: number = 30
	protected answerConfig: number = 15
	protected wagerConfig: number = 30
	protected firstRoundCategories: any[] = []
	protected secondRoundCategories: any[] = []
	protected searchResults: any[] = []
	protected categorySearch: string = ''
	protected searchLoader = false
	protected questionMode: string = 'cyo'
	private maxPlayers: number = 6

	constructor(
		private apiService: ApiService,
		private modal: ModalService,
		private jwtService: JwtService,
		private router: Router,
		private user: AuthService,
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
		this.botConfig = Math.min(Math.max(this.botConfig, 0), this.maxPlayers - 1)
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

	validateWagerConfig() {
		this.wagerConfig = Math.min(Math.max(this.wagerConfig, 3), 60)
	}

	hideConfig() {
		this.modal.hideConfig()
	}

	createPrivateGame(bots: number) {
		let playerImg = ''
		if (this.user.Authenticated()) {
			this.playerName = this.user.Name()
			playerImg = this.user.ImgUrl()
		}
		if (this.playerName == '') {
			document.getElementById('player-name-config')!.focus()
			document.getElementById('player-name-config')!.style.border = '1px solid red';
			setTimeout(() => {
				document.getElementById('player-name-config')!.style.border = '1px solid grey';
			}, 1000)
			return
		}
		this.apiService.CreatePrivateGame(
			this.playerName, playerImg,
			bots, this.twoRoundChecked, this.penaltyChecked,
			this.pickConfig, this.buzzConfig, this.answerConfig, this.wagerConfig,
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