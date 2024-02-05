import { Component, OnInit, AfterViewChecked } from '@angular/core';
import { Message } from '../../model/model';
import { PlayerService } from 'src/app/services/player.service';
import { JwtService } from 'src/app/services/jwt.service';
import { WebsocketService } from 'src/app/services/websocket.service';

@Component({
	selector: 'app-chat',
	templateUrl: './chat.component.html',
	styleUrls: ['./chat.component.less']
})
export class ChatComponent implements OnInit, AfterViewChecked {
	private jwt: string;
	protected messages: Message[] = [];
	protected message: string;
	protected hideChat = true;

	constructor(
		private websocketService: WebsocketService,
		protected player: PlayerService,
		protected jwtService: JwtService,
	) { }

	ngOnInit(): void {
		this.jwtService.jwt$.subscribe(jwt => {
			this.jwt = jwt;
		});
	}

	ngAfterViewChecked(): void {
		this.scrollToBottom();
	}

	sendMessage(): void {
		if (!this.message) {
			return;
		}
		this.messages.push({
			username: this.player.Name(),
			message: this.message,
			timestamp: new Date().toISOString(),
		});
		this.message = ''
	}

	scrollToBottom(): void {
		let chatMessages = document.getElementById('chat-messages')
		if (!chatMessages) {
			return;
		}
		chatMessages.scrollTop = chatMessages.scrollHeight;
	}

	openChat(): void {
		this.hideChat = false;
	}

	closeChat(): void {
		this.hideChat = true;
	}
}
