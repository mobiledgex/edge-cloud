import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import { Observable } from 'rxjs/Observable';

@Component({
  selector: 'app-clusters',
  templateUrl: './clusters.component.html',
  styleUrls: ['./clusters.component.scss']
})
export class ClustersComponent implements OnInit {
  clusters$: Object;
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getClusters().subscribe(
      (data) => (this.clusters$ = JSON.parse("[" + data.split('}\n{').join('},\n{') + "]"))
   );
  }

}
