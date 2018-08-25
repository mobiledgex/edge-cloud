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
      (data) => {
        data = "[" + data.split("}\n{").join("},\n{") + "]";
        this.appinsts$ = JSON.parse(data);
      }
    );
  }

//   ngOnInit() {
//     this.data.getAppInstances().subscribe(
//       (data) => {
//         // var data1 = '[' + data.replace('/\}\n\{/mg', '},\n{') + ']';
//         console.log(data);
//         var str1 = data;
//         var str2="[";
// for (var i=0; i < str1.length; i++) {
//      //console.log(str1.charAt(i));
//      str2 += str1.charAt(i);
//      if ( i < (str1.length -2) ) {
//              if (str1.charAt(i) === '}' && (str1.charAt(i+2) === '{')) {
//                   str2 += ',';
//              }
//      }
// }
//         str2 += ']';
//         console.log(str2);
//         var data2 = JSON.parse(str2);
        
//         this.appinsts$ = data2; 
//       }
//    );
//   }

}
