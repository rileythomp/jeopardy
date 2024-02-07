import { Component } from '@angular/core';
import { GameStateService } from 'src/app/services/game-state.service';
import { PlayerService } from 'src/app/services/player.service';

@Component({
  selector: 'app-question',
  templateUrl: './question.component.html',
  styleUrls: ['./question.component.less']
})
export class QuestionComponent {

  constructor(
    protected game: GameStateService,
    protected player: PlayerService
  ) { }
}
