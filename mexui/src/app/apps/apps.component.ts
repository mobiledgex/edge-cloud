import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import { Observable } from 'rxjs/Observable';
import * as _ from 'lodash';
import { map } from 'lodash';

@Component({
  selector: 'app-apps',
  templateUrl: './apps.component.html',
  styleUrls: ['./apps.component.scss']
})
export class AppsComponent implements OnInit {
  apps$: Object;
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getApps().subscribe(
      (data) => (
        this.apps$ = JSON.parse("[" + data.split('}\n{').join('},\n{') + "]")
      )
   );
  }
 
}
