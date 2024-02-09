import { Component, OnInit } from '@angular/core';
import { GameStateService } from 'src/app/services/game-state.service';
import { Question } from 'src/app/model/model';

@Component({
	selector: 'app-board-intro',
	templateUrl: './board-intro.component.html',
	styleUrls: ['./board-intro.component.less']
})
export class BoardIntroComponent implements OnInit {
	categories: string[];
	questionRows: Question[][];

	constructor(
		protected game: GameStateService,
	) {
		this.categories = this.game.Categories()
		this.questionRows = this.game.QuestionRows()
	}

	ngOnInit() {
		let questionValues = document.getElementsByClassName('question-value') as HTMLCollectionOf<HTMLElement>
		let indexes = Array.from({length: 30}, (v, i) => i) 
		let revealGroups: any[] = []
		for (let i = 0; i < 6; i++) {
			let revealGroup = []
			for (let j = 0; j < 5; j++) {
				let index = Math.floor(Math.random() * indexes.length);
			    let num = indexes.splice(index, 1)[0];
				revealGroup.push(num)
			}
			revealGroups.push(revealGroup)
		}
		let i = 0
		let valuesInterval = setInterval(() => {
			if (i < revealGroups.length) {
				for (let j = 0; j < revealGroups[i].length; j++) {
					questionValues[revealGroups[i][j]].style.color = 'var(--jeopardy-yellow)'
				}
				i++
			} else {
				clearInterval(valuesInterval)
			}
		}, 1000)
		setTimeout(() => {
			let categoryTitles = document.getElementsByClassName('category-title') as HTMLCollectionOf<HTMLElement>
			let j = 0
			let titlesInterval = setInterval(() => {
				if (j < categoryTitles.length) {
					(categoryTitles[j].firstChild as HTMLElement).classList.add('fade-in')
					j++
				} else {
					clearInterval(titlesInterval)
				}
			}, 2500)
		}, 5000)
	}
}
