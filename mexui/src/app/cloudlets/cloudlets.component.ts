import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import {Cloudlet} from '../app.model';

@Component({
  selector: 'app-cloudlets',
  templateUrl: './cloudlets.component.html',
  styleUrls: ['./cloudlets.component.scss']
})
export class CloudletsComponent implements OnInit {
  cloudlets$: Object;
  cloudlets: Cloudlet[] = [];  
  cldModel: Cloudlet;           
  showNew: Boolean = false;    
  submitType: string = 'Save';   
  selectedRow: number;   
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getCloudlets().subscribe(
      (data) => {
        // console.log(data);
        var obj = JSON.parse( "[" + data.split('}\n{').join('},\n{') + "]");
        var cldlist: Cloudlet[] = [];
        obj.forEach(function(entry) {
          cldlist.push(new Cloudlet(entry.result.key.operator_key.name, 
                                             entry.result.key.name,
                                             entry.result.access_uri,
                                             entry.result.location.lat,
                                             entry.result.location.long,
                                             entry.result.ip_support,
                                             entry.result.num_dynamic_ips
                                            ));
      });
        this.cloudlets$ = obj;
        this.cloudlets = cldlist;
        console.log(this.cloudlets);
      }
   );
  }
  
  onNew(){
    this.cldModel = new Cloudlet();
    this.submitType = 'Save';
    this.showNew = true;
 }

 onSave(){
   if (this.submitType === 'Save') {
     
     this.data.createCloudlet(this.cldModel).subscribe(
        (data) => {
         this.cloudlets.push(this.cldModel);
          // console.log(this.cldModel);
        }
     );
   } else {
     this.data.updateCloudlet(this.cldModel).subscribe(
       (data) => {
     // Update existing properties based on the model
     this.cloudlets[this.selectedRow].operatorName = this.cldModel.operatorName;
     this.cloudlets[this.selectedRow].cloudletName = this.cldModel.cloudletName;
     this.cloudlets[this.selectedRow].uri = this.cldModel.uri;
     this.cloudlets[this.selectedRow].locationLat  = this.cldModel.locationLat;
     this.cloudlets[this.selectedRow].locationLong  = this.cldModel.locationLong;
     this.cloudlets[this.selectedRow].ipSupport  = this.cldModel.ipSupport;
     this.cloudlets[this.selectedRow].numDynamicIps  = this.cldModel.numDynamicIps;
       }
     );
   }
   this.showNew = false;
 }

 
 onEdit(index: number){
     this.selectedRow = index;
     this.cldModel = new Cloudlet();
     this.cldModel = Object.assign({}, this.cloudlets[this.selectedRow]);
     this.submitType = 'Update';
     this.showNew = true;
 }

 onDelete(index: number) {
   this.selectedRow = index;
   this.cldModel = new Cloudlet();
   this.cldModel = Object.assign({}, this.cloudlets[this.selectedRow]);
   this.data.deleteCloudlet(this.cldModel).subscribe(
     (data) => {
         this.cloudlets.splice(index,1);
     }
   );
 }

 onCancel() {
   this.showNew = false;
 }
}
