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
import { RecvVoteComponent } from './game/recv-vote/recv-vote.component';
import { RecvWagerComponent } from './game/recv-wager/recv-wager.component';
import { PostGameComponent } from './game/post-game/post-game.component';
import { LinkJoinComponent } from './link-join/link-join.component';

@NgModule({
  declarations: [
    AppComponent,
    GameComponent,
    JoinComponent,
    PreGameComponent,
    RecvPickComponent,
    RecvBuzzComponent,
    RecvAnsComponent,
    RecvVoteComponent,
    RecvWagerComponent,
    PostGameComponent,
    LinkJoinComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    FormsModule,
    HttpClientModule,
  ],
  providers: [],
  bootstrap: [AppComponent],
  schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class AppModule { }
