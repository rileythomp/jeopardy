import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { GameComponent } from './game/game.component';
import { JoinComponent } from './join/join.component';
import { LobbyComponent } from './game/lobby/lobby.component';
import { BoardComponent } from './game/board/board.component';

@NgModule({
  declarations: [
    AppComponent,
    GameComponent,
    JoinComponent,
    LobbyComponent,
    BoardComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    FormsModule,
    HttpClientModule,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
