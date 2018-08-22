import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import { Observable } from 'rxjs/Observable';

@Component({
  selector: 'app-appinsts',
  templateUrl: './appinsts.component.html',
  styleUrls: ['./appinsts.component.scss']
})
export class AppinstsComponent implements OnInit {
  appinsts$: Object;
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getAppInstances().subscribe(
      (data) => (this.appinsts$ = data )
   );
  }

}
