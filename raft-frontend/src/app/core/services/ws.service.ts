import { Inject, Injectable } from '@angular/core';
import { Subject, Observable, BehaviorSubject } from 'rxjs';
import { webSocket } from 'rxjs/webSocket';

interface Message {
  serverId: string;
  type: string;
}

interface KeyUpdatedMessage extends Message {
  key: string;
  value: number;
}

interface ServerTypeChanged extends Message {
  serverType: number;
}

@Injectable({
  providedIn: 'root',
})
export class WebsocketService {
  heartbeatReceivedMessageType = 'HeatbeatReceivedMessage';
  serverTypeChangedMessageType = 'ServerTypeChangedMessage';
  keyUpdatedMessageType = 'KeyUpdatedMessage';

  public heartbeatReceived$: Subject<string> = new Subject();
  public keyUpdated$: Subject<KeyUpdatedMessage> = new Subject();
  public serverTypeChanged$: Subject<ServerTypeChanged> = new Subject();

  constructor() {}

  connect(serverId: string): void {
    let port = '';
    switch (serverId) {
      case 'Kim':
        port = '6969';
        break;
      case 'Ricky':
        port = '6970';
        break;
      case 'Laszlo':
        port = '6971';
        break;
      default:
        console.log('Invalid serverId');
        break;
    }

    if (!port) return;

    webSocket({
      url: `ws://localhost:${port}/ws`,
      deserializer: (e) => {
        return e.data;
      },
    }).subscribe(
      (msg) => this.messageHandler(msg), // Called whenever there is a message from the server.
      (err) => console.log(err), // Called if at any point WebSocket API signals some kind of error.
      () => console.log('complete') // Called when connection is closed (for whatever reason).
    );
  }

  private messageHandler(msg: any) {
    if (!msg) return;
    try {
      msg = JSON.parse(msg);
    } catch (error) {
      console.log("Couldn't deserialize:", msg, error);
    }
    const m: Message = msg;
    switch (m.type) {
      case this.heartbeatReceivedMessageType:
        this.heartbeatReceived$.next(m.serverId);
        break;
      case this.keyUpdatedMessageType:
        const keyUpadtedMessage = m as KeyUpdatedMessage;
        this.keyUpdated$.next({
          key: keyUpadtedMessage.key,
          serverId: keyUpadtedMessage.serverId,
          type: keyUpadtedMessage.type,
          value: keyUpadtedMessage.value,
        });
        break;
      case this.serverTypeChangedMessageType:
        const serverTypeChangedMessage = m as ServerTypeChanged;
        this.serverTypeChanged$.next({
          serverId: m.serverId,
          serverType: serverTypeChangedMessage.serverType,
          type: serverTypeChangedMessage.type,
        });
        break;
      default:
        throw 'unknown message';
    }
  }
}
