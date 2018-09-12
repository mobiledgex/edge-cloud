import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import {Flavor} from '../app.model';

@Component({
  selector: 'app-flavors',
  templateUrl: './flavors.component.html',
  styleUrls: ['./flavors.component.scss']
})
export class FlavorsComponent implements OnInit {
   flavors$: Object;
   flavors: Flavor[] = [];  
  flvModel: Flavor;           
  showNew: Boolean = false;    
  submitType: string = 'Save';   
  selectedRow: number;           
  
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getFlavors().subscribe(
      (data) => {
        var obj = JSON.parse("[" + data.split('}\n{').join('},\n{') + "]")
        var flvlist: Flavor[] = [];
        obj.forEach(function(entry) {
          flvlist.push(new Flavor(entry.result.key.name, 
                                             entry.result.ram,
                                             entry.result.vcpus,
                                             entry.result.disk
                                            ));
      });
        this.flavors$ = obj;
        this.flavors = flvlist;
      }
   );
  }

  onNew(){
    this.flvModel = new Flavor();
    this.submitType = 'Save';
    this.showNew = true;
 }

 onSave(){
   if (this.submitType === 'Save') {
     
     this.data.createFlavor(this.flvModel).subscribe(
        (data) => {
         this.flavors.push(this.flvModel);
          // console.log(this.flvModel);
        }
     );
   } else {
     this.data.updateFlavor(this.flvModel).subscribe(
       (data) => {
     // Update existing properties based on the model
     this.flavors[this.selectedRow].name = this.flvModel.name;
     this.flavors[this.selectedRow].ram = this.flvModel.ram;
     this.flavors[this.selectedRow].vcpus = this.flvModel.vcpus;
     this.flavors[this.selectedRow].disk  = this.flvModel.disk;
       }
     );
   }
   this.showNew = false;
 }

 
 onEdit(index: number){
     this.selectedRow = index;
     this.flvModel = new Flavor();
     this.flvModel = Object.assign({}, this.flavors[this.selectedRow]);
     this.submitType = 'Update';
     this.showNew = true;
 }

 onDelete(index: number) {
   this.selectedRow = index;
   this.flvModel = new Flavor();
   this.flvModel = Object.assign({}, this.flavors[this.selectedRow]);
   this.data.deleteFlavor(this.flvModel).subscribe(
     (data) => {
         this.flavors.splice(index,1);
     }
   );
 }

 onCancel() {
   this.showNew = false;
 }
}
