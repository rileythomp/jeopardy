import { CUSTOM_ELEMENTS_SCHEMA, NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { HttpClientModule } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { AnalyticsComponent } from './analytics/analytics.component';
import { AnswersComponent } from './answers/answers.component';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { AuthButtonComponent } from './auth/auth-button/auth-button.component';
import { LoginComponent } from './auth/login/login.component';
import { NewPasswordComponent } from './auth/new-password/new-password.component';
import { PasswordResetComponent } from './auth/password-reset/password-reset.component';
import { PwReqComponent } from './auth/pw-req/pw-req.component';
import { RegisterComponent } from './auth/register/register.component';
import { DisputeComponent } from './dispute/dispute.component';
import { GameMessageComponent } from './game-message/game-message.component';
import { BoardIntroComponent } from './game/board-intro/board-intro.component';
import { ChatComponent } from './game/chat/chat.component';
import { GameComponent } from './game/game.component';
import { PlayerPodiumComponent } from './game/player-podium/player-podium.component';
import { PostGameComponent } from './game/post-game/post-game.component';
import { PreGameComponent } from './game/pre-game/pre-game.component';
import { QuestionComponent } from './game/question/question.component';
import { RecvAnsComponent } from './game/recv-ans/recv-ans.component';
import { RecvBuzzComponent } from './game/recv-buzz/recv-buzz.component';
import { RecvPickComponent } from './game/recv-pick/recv-pick.component';
import { RecvWagerComponent } from './game/recv-wager/recv-wager.component';
import { InstructionsComponent } from './instructions/instructions.component';
import { ConfigComponent } from './join/config/config.component';
import { JoinComponent } from './join/join.component';
import { LeaderboardsComponent } from './leaderboards/leaderboards.component';
import { LinkJoinComponent } from './link-join/link-join.component';
import { ProfileComponent } from './profile/profile.component';
import { ReactionsComponent } from './reactions/reactions.component';
import { WarningComponent } from './warning/warning.component';

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
    PlayerPodiumComponent,
    AuthButtonComponent,
    RegisterComponent,
    PwReqComponent,
    LoginComponent,
    ProfileComponent,
    PasswordResetComponent,
    NewPasswordComponent,
    LeaderboardsComponent,
    ReactionsComponent,
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
