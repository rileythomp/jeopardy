import { Component } from '@angular/core';
import { Ping, Reaction } from '../model/model';
import { JwtService } from '../services/jwt.service';
import { PlayerService } from '../services/player.service';
import { ReactionsService } from '../services/reactions.service';

const emojisList = [
	[{ emoji: "👏", description: "clapping" }, { emoji: "🔥", description: "fire" }, { emoji: "😃", description: "smiley" }, { emoji: "😡", description: "angry" }, { emoji: "🤔", description: "thinking" }],
	[{ emoji: "😵", description: "dizzy" }, { emoji: "😐", description: "neutral" }, { emoji: "😤", description: "triumph" }, { emoji: "💸", description: "money" }, { emoji: "🎉", description: "party" }],
	[{ emoji: "😲", description: "shocked" }, { emoji: "🏆", description: "trophy" }, { emoji: "🧠", description: "brain" }, { emoji: "😢", description: "sad" }, { emoji: "😂", description: "laughing" }],
	[{ emoji: "💯", description: "hundred" }, { emoji: "🙃", description: "upside-down" }, { emoji: "😅", description: "sweat" }, { emoji: "😒", description: "unamused" }, { emoji: "😭", description: "crying" }],
	[{ emoji: "🙌", description: "raisedhands" }, { emoji: "💪", description: "strong" }, { emoji: "👎", description: "thumbsdown" }, { emoji: "👌", description: "ok" }, { emoji: "👍", description: "thumbsup" }],
	[{ emoji: "😎", description: "cool" }, { emoji: "📈", description: "chartup" }, { emoji: "📉", description: "chartdown" }, { emoji: "🖕", description: "fu" }, { emoji: "👽", description: "alien" }]
]

@Component({
	selector: 'app-reactions',
	templateUrl: './reactions.component.html',
	styleUrls: ['./reactions.component.less']
})
export class ReactionsComponent {
	protected emojisList = emojisList
	protected emojiFilter: string = ''
	protected reactionsList: Reaction[] = []
	protected reaction: string
	protected hideReactions = false
	protected unseenReactions = 0
	private goToBottom = false

	constructor(
		private reactions: ReactionsService,
		protected player: PlayerService,
		private jwt: JwtService,
	) { }

	ngOnInit(): void {
		this.reactions.Connect()

		this.reactions.OnOpen(() => {
			this.reactions.Send({ token: this.jwt.GetJWT() })
		})

		this.reactions.OnMessage((event: { data: string }) => {
			let resp = JSON.parse(event.data)

			if (resp.code >= 4400) {
				console.error(resp.reaction)
				return
			}

			if (resp.reaction == Ping) {
				return
			}

			this.reactionsList.push({
				username: resp.name,
				reaction: resp.reaction,
				timestamp: resp.timeStamp,
				randPos: resp.randPos,
			})

			if (this.hideReactions) {
				this.unseenReactions++
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

	sendReaction(emoji: string): void {
		this.reactions.Send({ reaction: emoji })
	}

	scrollToBottom(): void {
		let reactionsList = document.getElementById('reactions-list')
		if (!reactionsList) {
			return
		}
		reactionsList.scrollTop = reactionsList.scrollHeight
	}

	openReactions(): void {
		this.hideReactions = false
		this.unseenReactions = 0
		this.goToBottom = true
	}

	closeReactions(): void {
		this.hideReactions = true
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

	getRightPosition(randPos: number): string {
		return `calc(${randPos}px + var(--players-container-width))`
	}

	filterEmojis() {
		let flatEmojis = emojisList.flat()
		let filteredEmojis = flatEmojis.filter(emoji => emoji.description.toLowerCase().includes(this.emojiFilter.toLowerCase()))
		let groupedEmojis = []
		for (let i = 0; i < filteredEmojis.length; i += 5) {
			groupedEmojis.push(filteredEmojis.slice(i, i + 5))
		}
		if (groupedEmojis.length > 0 && groupedEmojis[groupedEmojis.length - 1].length < 5) {
			let emptyEmojisNeeded = 5 - groupedEmojis[groupedEmojis.length - 1].length;
			for (let i = 0; i < emptyEmojisNeeded; i++) {
				groupedEmojis[groupedEmojis.length - 1].push({ emoji: "", description: "" });
			}
		}
		this.emojisList = groupedEmojis
	}
}
