import { Component, OnInit, AfterViewChecked } from '@angular/core'
import { Message } from '../../model/model'
import { PlayerService } from 'src/app/services/player.service'
import { JwtService } from 'src/app/services/jwt.service'
import { ChatService } from 'src/app/services/chat.service'
import { Ping } from '../../model/model'


@Component({
	selector: 'app-chat',
	templateUrl: './chat.component.html',
	styleUrls: ['./chat.component.less']
})
export class ChatComponent implements OnInit, AfterViewChecked {
	private jwt: string
	protected messages: Message[] = []
	protected message: string
	protected hideChat = true
	protected unreadMessages = 0

	constructor(
		private chatService: ChatService,
		protected player: PlayerService,
		protected jwtService: JwtService,
	) { }

	ngOnInit(): void {
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt
		})

		this.chatService.Connect()

		this.chatService.OnOpen(() => {
			this.chatService.Send({ token: this.jwt })
		})

		this.chatService.OnMessage((event: { data: string }) => {
			let resp = JSON.parse(event.data)

			if (resp.code >= 4400) {
				// TODO: HANDLE THIS BETTER
				// alert(resp.message)
				console.log(resp.message)
				return
			}

			if (resp.message == Ping) {
				return
			}

			console.log(resp)

			this.messages.push({
				username: resp.playerName,
				message: resp.message,
				timestamp: resp.timeStamp,
			})

			if (this.hideChat) {
				this.unreadMessages++
			}
		})
	}

	ngAfterViewChecked(): void {
		this.scrollToBottom()
	}

	sendMessage(): void {
		if (!this.message) {
			return
		}
		this.chatService.Send({ message: this.message })
		this.message = ''
	}

	scrollToBottom(): void {
		let chatMessages = document.getElementById('chat-messages')
		if (!chatMessages) {
			return
		}
		chatMessages.scrollTop = chatMessages.scrollHeight
	}

	openChat(): void {
		this.hideChat = false
		this.unreadMessages = 0
	}

	closeChat(): void {
		this.hideChat = true
	}

	epochTo12HrFormat(epoch: number) {
		let date = new Date(epoch * 1000)
		let hours = date.getHours()
		let minutes = "0" + date.getMinutes()
		let suffix = hours >= 12 ? 'PM' : 'AM'
		hours = hours % 12
		hours = hours ? hours : 12
		return hours + ':' + minutes.slice(-2) + suffix
	}
}
