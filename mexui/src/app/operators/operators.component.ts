import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import { Observable } from 'rxjs/Observable';

@Component({
  selector: 'app-operators',
  templateUrl: './operators.component.html',
  styleUrls: ['./operators.component.scss']
})
export class OperatorsComponent implements OnInit {
  operators$: Object;
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getOperators().subscribe(
      (data) => (
        this.operators$ = JSON.parse("[" + data.split('}\n{').join('},\n{') + "]"))
   );
  }

}
