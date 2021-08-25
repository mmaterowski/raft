import { Inject, Injectable } from '@angular/core';
import { Subject, Observable, BehaviorSubject } from 'rxjs';
import { webSocket } from 'rxjs/webSocket';
@Injectable({
  providedIn: 'root',
})
export class WebsocketService {
  // Our socket connection
  public SocketSubject = new Subject<any>();
  public message$: BehaviorSubject<string> = new BehaviorSubject('');
  kimSocket: any;
  rickySocket: any;
  laszloSocket: any;

  constructor() {}

  connect(serverId: string): void {
    let port = '';
    switch (serverId) {
      case 'Kim':
        port = '6979';
        // this.kimSocket = io(`http://localhost:${port}/`);
        break;
      case 'Ricky':
        port = '6980';
        // this.rickySocket = io(`http://localhost:${port}/socket.io/socketio`);
        break;
      case 'Laszlo':
        port = '6981';
        const subject = webSocket({
          url: 'ws://localhost:6971/ws',
        });
        subject.subscribe(
          (msg) => console.log('message received: ', msg), // Called whenever there is a message from the server.
          (err) => console.log(err), // Called if at any point WebSocket API signals some kind of error.
          () => console.log('complete') // Called when connection is closed (for whatever reason).
        );
        // subject.next({ message: 'some message' });

        // subject.complete(); // Closes the connection.
        // subject.error({ code: 4000, reason: 'I think our app just broke!' });

        // this.laszloSocket = io(`http://localhost:6981`);
        break;
      default:
        break;
    }
  }
  public sendMessage(message: string) {
    debugger;
    // this.laszloSocket = io(`http://localhost:6981`);
    // this.rickySocket.emit('notice', message);
    // this.kimSocket.emit('notice', message);
    this.laszloSocket.emit('notice', message);
  }

  public getNewMessage = () => {
    this.rickySocket.on('notice', (message: string) => {
      this.message$.next(message);
    });

    this.laszloSocket.on('notice', (message: string) => {
      this.message$.next(message);
    });

    this.kimSocket.on('notice', (message: string) => {
      this.message$.next(message);
    });

    return this.message$.asObservable();
  };
  // of both an observer and observable.
}
