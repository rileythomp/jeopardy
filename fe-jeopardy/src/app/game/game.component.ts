import { Component, OnInit, ViewChild, ElementRef } from '@angular/core'
import { Router } from '@angular/router'
import { GameStateService } from '../services/game-state.service'
import { WebsocketService } from '../services/websocket.service'
import { PlayerService } from '../services/player.service'
import { JwtService } from '../services/jwt.service'
import { GameState, Ping } from '../model/model'
import { environment } from '../../environments/environment'

const pickTimeout = 60
const buzzTimeout = 60
const defaultAnsTimeout = 60
const dailyDoubleAnsTimeout = 60
const finalJeopardyAnsTimeout = 60
const voteTimeout = 60
const dailyDoubleWagerTimeout = 60
const finalJeopardyWagerTimeout = 60
const buzzDelay = 0

// const pickTimeout = 5
// const buzzTimeout = 5
// const defaultAnsTimeout = 10
// const dailyDoubleAnsTimeout = 10
// const finalJeopardyAnsTimeout = 10
// const voteTimeout = 5
// const dailyDoubleWagerTimeout = 10
// const finalJeopardyWagerTimeout = 10
// const buzzDelay = 2000 / 2

@Component({
	selector: 'app-game',
	templateUrl: './game.component.html',
	styleUrls: ['./game.component.less'],
})
export class GameComponent implements OnInit {
	private jwt: string
	private countdownInterval: NodeJS.Timeout
	protected gameLink: string
	protected countdownSeconds: number
	protected gameMessage: string
	protected questionAnswer: string
	protected wagerAmt: string

	@ViewChild('jeopardyAudio') private jeopardyAudio: ElementRef
	protected playMusic: boolean = false
	protected showMusic: boolean = true
	protected showMusicInfo: boolean = false

	constructor(
		private router: Router,
		private websocketService: WebsocketService,
		private jwtService: JwtService,
		protected game: GameStateService,
		protected player: PlayerService,
	) {
		this.gameLink = environment.gameLink
	}

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
				// TODO: REPLACE WITH MODAL
				alert(resp.message)
				if (resp.code == 4500 || resp.code == 4401) {
					this.router.navigate(['/join'])
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
				alert(resp.message)
				return
			}

			if (this.game.IsPaused()) {
				this.countdownSeconds = 0
				clearInterval(this.countdownInterval)
				// TODO: REPLACE WITH MODAL
				alert(`${resp.message}, will resume when 3 players are ready`)
				return
			}

			switch (this.game.State()) {
				case GameState.PreGame:
				case GameState.PostGame:
					break
				case GameState.RecvPick:
					this.stopMusic()
					this.showMusic = false
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
						this.showMusic = true
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
						this.showMusic = true
						this.startCountdownTimer(finalJeopardyWagerTimeout)
					}
					break
				default:
					// TODO: REPLACE WITH MODAL
					alert('Unable to update game, redirecting to home page')
					this.router.navigate(['/join'])
					break
			}
		})
	}

	startCountdownTimer(seconds: number): void {
		clearInterval(this.countdownInterval)
		this.countdownSeconds = seconds
		this.countdownInterval = setInterval(() => {
			this.countdownSeconds -= 1
			if (this.countdownSeconds <= 0) {
				clearInterval(this.countdownInterval)
			}
		}, 1000)
	}

	openJoinLink(): void {
		window.open('join/' + this.game.Name(), '_blank')
	}

	copyJoinLink(): void {
		let joinLink = `${this.gameLink}/join/${this.game.Name()}`
		navigator.clipboard.writeText(joinLink).then(function () { }, function (err) { })
	}

	startMusic(): void {
		this.playMusic = true
		this.jeopardyAudio.nativeElement.play()
	}

	stopMusic(): void {
		this.playMusic = false
		this.jeopardyAudio.nativeElement.pause()
	}
}
