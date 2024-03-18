import { NgModule, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { GameComponent } from './game/game.component';
import { JoinComponent } from './join/join.component';
import { PreGameComponent } from './game/pre-game/pre-game.component';
import { RecvPickComponent } from './game/recv-pick/recv-pick.component';
import { RecvBuzzComponent } from './game/recv-buzz/recv-buzz.component';
import { RecvAnsComponent } from './game/recv-ans/recv-ans.component';
import { RecvWagerComponent } from './game/recv-wager/recv-wager.component';
import { PostGameComponent } from './game/post-game/post-game.component';
import { LinkJoinComponent } from './link-join/link-join.component';
import { ChatComponent } from './game/chat/chat.component';
import { QuestionComponent } from './game/question/question.component';
import { BoardIntroComponent } from './game/board-intro/board-intro.component';
import { WarningComponent } from './warning/warning.component';
import { AnalyticsComponent } from './analytics/analytics.component';
import { InstructionsComponent } from './instructions/instructions.component';
import { DisputeComponent } from './dispute/dispute.component';
import { GameMessageComponent } from './game-message/game-message.component';
import { ConfigComponent } from './join/config/config.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { AnswersComponent } from './answers/answers.component';

@NgModule({
  declarations: [
    AppComponent,
    GameComponent,
    JoinComponent,
    PreGameComponent,
    RecvPickComponent,
    RecvBuzzComponent,
    RecvAnsComponent,
    RecvWagerComponent,
    PostGameComponent,
    LinkJoinComponent,
    ChatComponent,
    QuestionComponent,
    BoardIntroComponent,
    WarningComponent,
    AnalyticsComponent,
    InstructionsComponent,
    DisputeComponent,
    GameMessageComponent,
    ConfigComponent,
    AnswersComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    FormsModule,
    HttpClientModule,
    BrowserAnimationsModule,
  ],
  providers: [],
  bootstrap: [AppComponent],
  schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class AppModule { }
