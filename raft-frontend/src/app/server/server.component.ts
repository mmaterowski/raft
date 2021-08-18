import { Component, Input, OnInit } from '@angular/core';
import { Circle } from '../servers/circle';

@Component({
  selector: 'g[app-server]',
  templateUrl: './server.component.html',
  styleUrls: ['./server.component.scss'],
})
export class ServerComponent implements OnInit {
  @Input()
  public circle: Circle | undefined;

  constructor() {}

  ngOnInit(): void {}
}
