import { Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import * as d3 from '../custom-d3';
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
  public laszlo: Circle | undefined;
  public kim: Circle | undefined;

  constructor() {}

  ngOnInit(): void {
    this.setupPanAndZoom();
    this.ricky = new Circle(300, 200, 20, 'chartreuse', 'Ricky');
    this.kim = new Circle(200, 100, 20, 'chartreuse', 'Kim');
    this.laszlo = new Circle(400, 100, 20, 'chartreuse', 'Laszlo');
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
