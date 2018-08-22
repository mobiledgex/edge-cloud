import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';

@Component({
  selector: 'app-flavors',
  templateUrl: './flavors.component.html',
  styleUrls: ['./flavors.component.scss']
})
export class FlavorsComponent implements OnInit {
   flavors$: Object;
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getFlavors().subscribe(
      (data) => (this.flavors$ = data )
   );
  }

}
