import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';

const jeopardyJWT = 'jeopardyJWT';

@Injectable({
  providedIn: 'root'
})
export class JwtService {
  private jwtSubject: BehaviorSubject<string>;
  public jwt$: Observable<string>;

  constructor() {
    const storedJwt: string = localStorage.getItem(jeopardyJWT) ?? '';
    this.jwtSubject = new BehaviorSubject<string>(storedJwt);
    this.jwt$ = this.jwtSubject.asObservable();
  }

  SetJWT(jwt: string): void {
    localStorage.setItem(jeopardyJWT, jwt);
    this.jwtSubject.next(jwt);
  }

  GetJWT(): string {
    return localStorage.getItem(jeopardyJWT) ?? '';
  }
}
