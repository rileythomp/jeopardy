import { AfterViewChecked, Component, OnInit } from '@angular/core'
import { ChatService } from 'src/app/services/chat.service'
import { JwtService } from 'src/app/services/jwt.service'
import { PlayerService } from 'src/app/services/player.service'
import { Message, Ping } from '../../model/model'


@Component({
	selector: 'app-chat',
	templateUrl: './chat.component.html',
	styleUrls: ['./chat.component.less']
})
export class ChatComponent implements OnInit, AfterViewChecked {
	protected messages: Message[] = []
	protected message: string
	protected hideChat = true
	protected unreadMessages = 0
	private goToBottom = false

	constructor(
		private chat: ChatService,
		protected player: PlayerService,
		private jwt: JwtService,
	) { }

	ngOnInit(): void {
		this.chat.Connect()

		this.chat.OnOpen(() => {
			this.chat.Send({ token: this.jwt.GetJWT() })
		})

		this.chat.OnMessage((event: { data: string }) => {
			let resp = JSON.parse(event.data)

			if (resp.code >= 4400) {
				console.error(resp.message)
				this.messages.push({
					username: 'Jeopardy System',
					message: resp.message,
					timestamp: resp.timeStamp,
				})
				return
			}

			if (resp.message == Ping) {
				return
			}

			this.messages.push({
				username: resp.name,
				message: resp.message,
				timestamp: resp.timeStamp,
			})

			if (this.hideChat) {
				this.unreadMessages++
			}

			this.goToBottom = true
		})
	}

	ngAfterViewChecked(): void {
		if (this.goToBottom) {
			this.scrollToBottom()
			this.goToBottom = false
		}
	}

	sendMessage(): void {
		if (!this.message) {
			return
		}
		this.chat.Send({ message: this.message })
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
		this.goToBottom = true
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
