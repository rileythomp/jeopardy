import { Component, OnInit, ViewChild, ElementRef } from '@angular/core'
import { Router } from '@angular/router'
import { GameStateService } from '../services/game-state.service'
import { WebsocketService } from '../services/websocket.service'
import { PlayerService } from '../services/player.service'
import { JwtService } from '../services/jwt.service'
import { GameState, Ping } from '../model/model'
import { ModalComponent } from '../modal/modal.component'

const pickTimeout = 30
const buzzTimeout = 30
const defaultAnsTimeout = 30
const dailyDoubleAnsTimeout = 30
const finalJeopardyAnsTimeout = 30
const voteTimeout = 10
const dailyDoubleWagerTimeout = 30
const finalJeopardyWagerTimeout = 30
const buzzDelay = 0

@Component({
	selector: 'app-game',
	templateUrl: './game.component.html',
	styleUrls: ['./game.component.less'],
})
export class GameComponent implements OnInit {
	private jwt: string
	private countdownInterval: NodeJS.Timeout
	protected gameLink: string
	protected gameMessage: string
	protected questionAnswer: string
	protected wagerAmt: string

	@ViewChild('jeopardyAudio') private jeopardyAudio: ElementRef
	protected playMusic: boolean = false
	protected showMusicInfo: boolean = false

	@ViewChild(ModalComponent) private modal: ModalComponent

	constructor(
		private router: Router,
		private websocketService: WebsocketService,
		private jwtService: JwtService,
		protected game: GameStateService,
		protected player: PlayerService,
	) { }

	ngOnInit(): void {
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt
		})

		let showJeopardyMusicInfo = localStorage.getItem('showJeopardyMusicInfo')
		if (showJeopardyMusicInfo === null) {
			this.showMusicInfo = true
			setTimeout(() => {
				this.showMusicInfo = false
			}, 5000)
			localStorage.setItem('showJeopardyMusicInfo', 'shown')
		}

		this.websocketService.Connect('play')

		this.websocketService.OnOpen(() => {
			let playReq = {
				token: this.jwt,
			}
			this.websocketService.Send(playReq)
		})

		this.websocketService.OnMessage((event: { data: string }) => {
			let resp = JSON.parse(event.data)

			if (resp.code >= 4400) {
				switch (resp.code) {
					case 4401:
					case 4500:
						this.modal.showMessage(resp.message)
						break
					case 4400:
						this.modal.showMessage(resp.message)
						break
				}
				return
			}

			if (resp.message == Ping) {
				return
			}

			console.log(resp)

			this.game.updateGameState(resp.game)
			this.player.updatePlayer(resp.curPlayer)
			this.gameMessage = resp.message

			if (resp.code == 4100) {
				this.modal.showMessage(resp.message)
				return
			}

			if (this.game.IsPaused()) {
				this.cancelCountdown()
				this.modal.showMessage(`${resp.message}, will resume when 3 players are ready`)
				return
			}

			switch (this.game.State()) {
				case GameState.PreGame:
				case GameState.PostGame:
				case GameState.BoardIntro:
					this.cancelCountdown()
					break
				case GameState.RecvPick:
					this.startCountdownTimer(pickTimeout)
					break
				case GameState.RecvBuzz:
					if (this.game.CurQuestionFirstBuzz()) {
						this.game.BlockBuzz(true)
						setTimeout(() => {
							this.game.BlockBuzz(false)
							if (this.game.StartBuzzCountdown()) {
								this.startCountdownTimer(buzzTimeout - buzzDelay / 1000)
							}
						}, buzzDelay)
					} else {
						if (this.game.StartBuzzCountdown()) {
							this.startCountdownTimer(buzzTimeout)
						}
					}
					break
				case GameState.RecvAns:
					if (!this.game.FinalRound()) {
						this.startCountdownTimer(defaultAnsTimeout)
					} else if (this.game.StartFinalAnswerCountdown()) {
						this.startCountdownTimer(finalJeopardyAnsTimeout)
					}
					break
				case GameState.RecvVote:
					if (this.player.CanVote()) {
						this.startCountdownTimer(voteTimeout)
					}
					break
				case GameState.RecvWager:
					if (!this.game.FinalRound()) {
						this.startCountdownTimer(dailyDoubleWagerTimeout)
					} else if (this.game.StartFinalWagerCountdown()) {
						this.startCountdownTimer(finalJeopardyWagerTimeout)
					}
					break
				default:
					this.modal.showMessage('Error while updating game')
					break
			}
		})
	}

	startCountdownTimer(seconds: number): void {
		this.cancelCountdown()
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

	abs(num: number): number {
		if (!num) {
			return 0
		}
		return Math.abs(num)
	}
}
