import { animate, state, style, transition, trigger } from '@angular/animations'
import { Component, ElementRef, OnInit, ViewChild } from '@angular/core'
import { ActivatedRoute, Router } from '@angular/router'
import { GameState, Ping, Player } from '../model/model'
import { GameStateService } from '../services/game-state.service'
import { JwtService } from '../services/jwt.service'
import { ModalService } from '../services/modal.service'
import { PlayerService } from '../services/player.service'
import { WebsocketService } from '../services/websocket.service'

const BadRequest = 4400
const Unauthorized = 4401
const ServerError = 4500

function isError(code: number): boolean {
	return code >= 4400
}

@Component({
	selector: 'app-game',
	templateUrl: './game.component.html',
	styleUrls: ['./game.component.less'],
	animations: [
		trigger('boardIntroFade', [
			state('void', style({ opacity: 0 })),
			transition('void => *', animate(2000)),
		]),
	]
})
export class GameComponent implements OnInit {
	private countdownInterval: NodeJS.Timeout
	protected joinPath: string
	protected gameMessage: string
	protected questionAnswer: string
	protected wagerAmt: string
	protected scoreChanges: any = {}

	showPauseGame: boolean = false

	@ViewChild('jeopardyAudio') private jeopardyAudio: ElementRef
	protected playMusic: boolean = false
	protected showMusicInfo: boolean = false

	constructor(
		private websocket: WebsocketService,
		private jwt: JwtService,
		protected game: GameStateService,
		protected player: PlayerService,
		private modal: ModalService,
		private router: Router,
		private route: ActivatedRoute,
	) { }

	ngOnInit(): void {
		this.route.paramMap.subscribe(params => {
			let joinCode = params.get('joinCode')
			this.joinPath = joinCode ? `/join/${joinCode}` : '/'
			this.websocket.Connect(`play/${joinCode}`)
		})
		let showJeopardyMusicInfo = localStorage.getItem('showJeopardyMusicInfo')
		if (showJeopardyMusicInfo === null) {
			this.showMusicInfo = true
			setTimeout(() => {
				this.showMusicInfo = false
			}, 5000)
			localStorage.setItem('showJeopardyMusicInfo', 'shown')
		}

		let showPauseGame = localStorage.getItem('showPauseGame')
		if (showPauseGame === null) {
			this.showPauseGame = true
			setTimeout(() => {
				this.showPauseGame = false
			}, 5000)
			localStorage.setItem('showPauseGame', 'shown')
		}

		this.websocket.OnOpen(() => {
			this.websocket.Send({
				state: GameState.PreGame,
				token: this.jwt.GetJWT(),
			})
		})

		this.websocket.OnMessage((event: { data: string }) => {
			let resp = JSON.parse(event.data)

			if (isError(resp.code)) {
				switch (resp.code) {
					case BadRequest:
					case Unauthorized:
						this.router.navigate([this.joinPath])
						break
					case ServerError:
						this.modal.displayMessage(resp.message)
						break
				}
				return
			}

			if (resp.message == Ping) {
				return
			}

			let savedPlayers = this.game.Players()

			this.game.updateGameState(resp.game)
			this.player.updatePlayer(resp.curPlayer)
			this.gameMessage = resp.message

			if (resp.code == 4100) {
				this.modal.displayMessage(resp.message)
				return
			}

			if (this.game.IsPaused()) {
				this.cancelCountdown()
				return
			}

			this.handleScoreChanges(savedPlayers)

			switch (this.game.State()) {
				case GameState.PreGame:
				case GameState.PostGame:
				case GameState.BoardIntro:
					this.cancelCountdown()
					break
				case GameState.RecvDispute:
					this.cancelCountdown()
					this.modal.displayDispute()
					break
				case GameState.RecvPick:
					this.startCountdownTimer(this.game.PickTimeout())
					break
				case GameState.RecvBuzz:
					this.cancelCountdown()
					if (this.game.CurQuestionFirstBuzz()) {
						this.game.BlockBuzz(true)
						let buzzDelay = this.game.BuzzDelay()
						setTimeout(() => {
							this.game.BlockBuzz(false)
							this.startCountdownTimer(this.game.BuzzTimeout() - buzzDelay)
						}, buzzDelay * 1000)
					} else {
						this.startCountdownTimer(this.game.BuzzTimeout())
					}
					break
				case GameState.RecvAns:
					if (!this.game.FinalRound()) {
						this.startCountdownTimer(this.game.AnswerTimeout())
					} else if (this.game.StartFinalAnswerCountdown()) {
						this.startCountdownTimer(this.game.FinalAnswerTimeout())
					}
					break
				case GameState.RecvWager:
					if (!this.game.FinalRound()) {
						this.startCountdownTimer(this.game.WagerTimeout())
					} else if (this.game.StartFinalWagerCountdown()) {
						this.startCountdownTimer(this.game.FinalWagerTimeout())
					}
					break
				default:
					this.modal.displayMessage('Error while updating game')
					break
			}
		})
	}

	startCountdownTimer(seconds: number): void {
		this.cancelCountdown()
		if (this.game.IsPaused()) {
			return
		}
		let countdownBar = document.getElementById('countdown-bar')
		for (let i = 0; i < 2 * (seconds - 1); i++) {
			let countdownBox = document.createElement('div')
			countdownBox.id = `countdown-${i}`
			countdownBox.style.backgroundColor = 'red'
			countdownBar?.appendChild(countdownBox)
		}
		let start = 0
		let end = countdownBar!.children.length - 1
		this.countdownInterval = setInterval(() => {
			document.getElementById(`countdown-${start}`)!.style.backgroundColor = 'white'
			document.getElementById(`countdown-${end}`)!.style.backgroundColor = 'white'
			start += 1
			end -= 1
		}, 1000)
	}

	cancelCountdown(): void {
		clearInterval(this.countdownInterval)
		let countdownBar = document.getElementById('countdown-bar')
		while (countdownBar?.firstChild) {
			countdownBar.removeChild(countdownBar.firstChild)
		}
	}

	startMusic(): void {
		this.playMusic = true
		this.jeopardyAudio.nativeElement.play()
	}

	stopMusic(): void {
		this.playMusic = false
		this.jeopardyAudio.nativeElement.pause()
	}

	pauseGame(pause: boolean) {
		this.modal.displayMessage(`Game ${pause ? 'paused' : 'resumed'}`)
		this.websocket.Send({
			state: this.game.State(),
			pause: pause ? 1 : -1,
		})
	}

	handleScoreChanges(savedPlayers: Player[]): void {
		for (let i = 0; i < savedPlayers.length; i++) {
			let savedPlayer = savedPlayers[i]
			for (let j = 0; j < this.game.Players().length; j++) {
				let curPlayer = this.game.Players()[j]
				if (savedPlayer.id == curPlayer.id) {
					this.scoreChanges[curPlayer.id] = curPlayer.score - savedPlayer.score
				}
			}
		}

		setTimeout(() => {
			this.scoreChanges = {}
		}, 3000)
	}
}
