import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { GameComponent } from './game/game.component';
import { LobbyComponent } from './lobby/lobby.component';
import { JoinComponent } from './join/join.component';

@NgModule({
  declarations: [
    AppComponent,
    GameComponent,
    LobbyComponent,
    JoinComponent,
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
