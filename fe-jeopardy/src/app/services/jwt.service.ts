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
    const storedJwt: string = sessionStorage.getItem(jeopardyJWT) ?? '';
    this.jwtSubject = new BehaviorSubject<string>(storedJwt);
    this.jwt$ = this.jwtSubject.asObservable();
  }

  SetJWT(jwt: string): void {
    sessionStorage.setItem(jeopardyJWT, jwt);
    this.jwtSubject.next(jwt);
  }

  GetJWT(): string {
    return sessionStorage.getItem(jeopardyJWT) ?? '';
  }
}
