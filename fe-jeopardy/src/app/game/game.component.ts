import { Component, OnInit, ViewChild, ElementRef } from '@angular/core'
import { GameStateService } from '../services/game-state.service'
import { WebsocketService } from '../services/websocket.service'
import { PlayerService } from '../services/player.service'
import { JwtService } from '../services/jwt.service'
import { GameState, Ping, Player } from '../model/model'
import { ModalService } from '../services/modal.service'
import { trigger, state, style, animate, transition } from '@angular/animations';

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
	private jwt: string
	private countdownInterval: NodeJS.Timeout
	protected gameLink: string
	protected gameMessage: string
	protected questionAnswer: string
	protected wagerAmt: string
	protected scoreId: string
	protected scoreChange: number = 0

	showPauseGame: boolean = false

	@ViewChild('jeopardyAudio') private jeopardyAudio: ElementRef
	protected playMusic: boolean = false
	protected showMusicInfo: boolean = false

	constructor(
		private websocketService: WebsocketService,
		private jwtService: JwtService,
		protected game: GameStateService,
		protected player: PlayerService,
		private modal: ModalService,
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

		let showPauseGame = localStorage.getItem('showPauseGame')
		if (showPauseGame === null) {
			this.showPauseGame = true
			setTimeout(() => {
				this.showPauseGame = false
			}, 5000)
			localStorage.setItem('showPauseGame', 'shown')
		}

		this.websocketService.Connect('play')

		this.websocketService.OnOpen(() => {
			this.websocketService.Send({
				state: GameState.PreGame,
				token: this.jwt,
			})
		})

		this.websocketService.OnMessage((event: { data: string }) => {
			let resp = JSON.parse(event.data)

			if (resp.code >= 4400) {
				switch (resp.code) {
					case 4400:
					case 4401:
					case 4500:
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

			console.log(resp)

			if (resp.code == 4100) {
				this.modal.displayMessage(resp.message)
				return
			}

			if (this.game.IsPaused()) {
				this.cancelCountdown()
				return
			}

scoreChangeLoop:
			for (let i = 0; i < savedPlayers.length; i++) {
				let savedPlayer = savedPlayers[i]
				for (let j = 0; j < this.game.Players().length; j++) {
					let curPlayer = this.game.Players()[j]
					if (savedPlayer.id != curPlayer.id) {
						continue
					}
					this.scoreId = curPlayer.id
					this.scoreChange = curPlayer.score - savedPlayer.score
					if (this.scoreChange != 0) {
						setTimeout(() => {
							this.scoreChange = 0
						}, 3000)
						break scoreChangeLoop
					}
					break
				}
			}

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
							if (this.game.StartBuzzCountdown()) {
								this.startCountdownTimer(this.game.BuzzTimeout() - buzzDelay)
							}
						}, buzzDelay * 1000)
					} else if (this.game.StartBuzzCountdown()) {
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

	abs(num: number): number {
		if (!num) {
			return 0
		}
		return Math.abs(num)
	}

	pauseGame(pause: boolean) {
		this.modal.displayMessage(`Game ${pause ? 'paused' : 'resumed'}`)
		this.websocketService.Send({
			state: this.game.State(),
			pause: pause ? 1 : -1,
		})
	}

	onClock(player: Player): boolean {
		return player.canPick || player.canAnswer || player.canWager
	}
}
