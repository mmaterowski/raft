import { Component, Input, OnInit } from '@angular/core';
import { Line } from '../model/line';
import { Point } from '../model/point';

type NewType = Line;

@Component({
  selector: 'g[app-heartbeat]',
  templateUrl: './heartbeat.component.html',
  styleUrls: ['./heartbeat.component.scss'],
})
export class HeartbeatComponent implements OnInit {
  constructor() {}
  @Input()
  public line: Line | undefined;

  @Input()
  public receivesHearbeat: boolean | undefined;

  public midPoint: Point | undefined;
  ngOnInit(): void {
    if (!this.line) throw Error('No line provided');
    this.midPoint = new Point(0, 0);
    this.midPoint.X = (this.line.start.X + this.line.end.X) / 2;
    this.midPoint.Y = (this.line.start.Y + this.line.end.Y) / 2;
  }
}
