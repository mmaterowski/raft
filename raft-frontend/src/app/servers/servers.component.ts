import { Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import { WebsocketService } from '../core/services/ws.service';
import * as d3 from '../custom-d3';
import { Line } from '../model/line';
import { Point } from '../model/point';
import { Circle } from './circle';

@Component({
  selector: 'app-servers',
  templateUrl: './servers.component.html',
  styleUrls: ['./servers.component.scss'],
})
export class ServersComponent implements OnInit {
  @ViewChild('canvasContainer', { static: true })
  canvasContainer: ElementRef<SVGSVGElement> | undefined;
  @ViewChild('canvasViewBox', { static: true })
  canvasViewBox: ElementRef | undefined;
  public ricky: Circle | undefined;
  public rickyToLaszloHeart: Line | undefined;
  public rickyToKimHeart: Line | undefined;
  public kimToLaszloHeart: Line | undefined;
  public laszlo: Circle | undefined;
  public kim: Circle | undefined;
  public kimEntries: { key: string; value: number }[] = [
    { key: 'sx34xfsdf', value: 12 },
    { key: 'sxxxscsdf', value: 12 },
    { key: 'sxxxxxfssdfdf', value: 52 },
  ];
  public laszloEntries: { key: string; value: number }[] = [
    { key: 'kjl', value: 182 },
    { key: 'lkj', value: 12 },
    { key: 'sxxxxxfssdfdf', value: 22 },
  ];

  public rickyEntries: { key: string; value: number }[] = [
    { key: 'jeba', value: 32225 },
    { key: 'c', value: 1 },
    { key: 'telewije', value: 334 },
    { key: 'polska', value: 53 },
  ];
  constructor(private wsService: WebsocketService) {}

  ngOnInit(): void {
    this.setupPanAndZoom();
    this.prepareHeartbeats();
    this.wsService.connect('Kim');
    this.wsService.connect('Laszlo');
    this.wsService.connect('Ricky');

    this.wsService.heartbeatReceived$.subscribe((r) => console.log(r));
    this.wsService.keyUpdated$.subscribe((r) => {
      if (r.serverId == 'Kim') {
        this.kimEntries.push({ key: r.key, value: r.value });
        this.kimEntries = this.kimEntries.reverse();
      }

      if (r.serverId == 'Ricky') {
        this.rickyEntries.push({ key: r.key, value: r.value });
        this.rickyEntries = this.rickyEntries.reverse();
      }

      if (r.serverId == 'Laszlo') {
        this.laszloEntries.push({ key: r.key, value: r.value });
        this.laszloEntries = this.laszloEntries.reverse();
      }
    });
    this.wsService.serverTypeChanged$.subscribe((r) => console.log(r));
    // this.signalR.addHeartbeatListener();
    // this.signalR.addServerTypeChangedListener();
    // this.signalR.addStateUpdateListener();
  }

  private prepareHeartbeats() {
    this.ricky = new Circle(300, 200, 20, 'chartreuse', 'Ricky');
    this.kim = new Circle(200, 100, 20, 'chartreuse', 'Kim');
    this.laszlo = new Circle(400, 100, 20, 'chartreuse', 'Laszlo');

    this.rickyToLaszloHeart = new Line(
      { X: this.ricky.x, Y: this.ricky.y },
      { X: this.laszlo.x, Y: this.laszlo.y }
    );

    this.rickyToKimHeart = new Line(
      { X: this.ricky.x, Y: this.ricky.y },
      { X: this.kim.x, Y: this.kim.y }
    );

    this.kimToLaszloHeart = new Line(
      { X: this.kim.x, Y: this.kim.y },
      { X: this.laszlo.x, Y: this.laszlo.y }
    );
  }

  private setupPanAndZoom() {
    const d3ElemContainer = d3.select(
      this.canvasContainer?.nativeElement as any
    );
    const d3ElemViewBoc = d3.select(this.canvasViewBox?.nativeElement);

    d3ElemContainer.call(
      d3
        .zoom()
        .on('zoom', (event: { transform: never }) => {
          d3ElemViewBoc.attr('transform', event.transform);
        })
        .filter((event) => event.button === 1 || event.type === 'wheel')
    );
  }
}
