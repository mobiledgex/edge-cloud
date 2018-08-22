import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import { Observable } from 'rxjs/Observable';

@Component({
  selector: 'app-cloudlets',
  templateUrl: './cloudlets.component.html',
  styleUrls: ['./cloudlets.component.scss']
})
export class CloudletsComponent implements OnInit {
  cloudlets$: Object;
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getCloudlets().subscribe(
      (data) => (this.cloudlets$ = data )
   );
  }

}
