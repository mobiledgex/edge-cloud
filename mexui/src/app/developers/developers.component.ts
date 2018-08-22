import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import { Observable } from 'rxjs/Observable';

@Component({
  selector: 'app-developers',
  templateUrl: './developers.component.html',
  styleUrls: ['./developers.component.scss']
})
export class DevelopersComponent implements OnInit {
  developers$: Object;
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getDevelopers().subscribe(
       (data) => (this.developers$ = data )
    );
  }

}
