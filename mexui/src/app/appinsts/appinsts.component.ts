import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import {AppInst} from '../app.model';


@Component({
  selector: 'app-appinsts',
  templateUrl: './appinsts.component.html',
  styleUrls: ['./appinsts.component.scss']
})

export class AppinstsComponent implements OnInit {
  appinsts$: Object;

  appinsts: AppInst[] = [];
  instModel: AppInst;
  showNew: Boolean = false;
  submitType: string = 'Save';
  selectedRow: number;

  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getAppInstances().subscribe(
      (data) => {
        data = "[" + data.split("}\n{").join("},\n{") + "]";
        var obj = JSON.parse(data);
        var instlist: AppInst[] = [];
        obj.forEach(function(entry){
          instlist.push(new AppInst(
                entry.result.key.id,
                entry.result.key.app_key.developer_key.name,
                entry.result.key.app_key.name,
                entry.result.key.app_key.version,
                entry.result.image_path,
                entry.result.image_type,
                entry.result.uri,
                entry.result.liveness,
                entry.result.mapped_ports,
                entry.result.mapped_path,
                entry.result.flavor.name,
                entry.result.key.cloudlet_key.operator_key.name,
                entry.result.key.cloudlet_key.name,
                entry.result.cloudlet_loc.lat,
                entry.result.cloudlet_loc.long,
                entry.result.cluster_inst_key.cluster_key.name,
                entry.result.access_layer,
          )
         )
       });
        this.appinsts$ = obj;
        this.appinsts = instlist;
      }
    );
  }

  onNew(){
    this.instModel = new AppInst();
    this.submitType = 'Save';
    this.showNew = true;
 }

 onSave(){
   if (this.submitType === 'Save') {
     
     this.data.createAppInst(this.instModel).subscribe(
        (data) => {
         this.appinsts.push(this.instModel);
          // console.log(this.instModel);
        },
      (error) => {
        // console.log ("We should not get here!");
        console.log(this.instModel);
        this.appinsts.push(this.instModel);
      }
     );
   } else {
     this.data.updateAppInst(this.instModel).subscribe(
       (data) => {
     // Update existing properties based on the model
     this.appinsts[this.selectedRow].id = this.instModel.id;
     this.appinsts[this.selectedRow].developerName = this.instModel.developerName;
     this.appinsts[this.selectedRow].name = this.instModel.name;
     this.appinsts[this.selectedRow].version = this.instModel.version;
     this.appinsts[this.selectedRow].imagePath  = this.instModel.imagePath;
     this.appinsts[this.selectedRow].imageType = this.instModel.imageType;
     this.appinsts[this.selectedRow].uri = this.instModel.uri;
     this.appinsts[this.selectedRow].liveNess = this.instModel.liveNess;
     this.appinsts[this.selectedRow].mappedPorts = this.instModel.mappedPorts;
     this.appinsts[this.selectedRow].mappedPath = this.instModel.mappedPath;
     this.appinsts[this.selectedRow].flavor = this.instModel.flavor;
     this.appinsts[this.selectedRow].cloudletKey = this.instModel.cloudletKey;
     this.appinsts[this.selectedRow].cloudletName = this.instModel.cloudletName;
     this.appinsts[this.selectedRow].cloudletLat = this.instModel.cloudletLat;
     this.appinsts[this.selectedRow].cloudletLong = this.instModel.cloudletLong;
     this.appinsts[this.selectedRow].cluster = this.instModel.cluster;
     this.appinsts[this.selectedRow].accessLayer = this.instModel.accessLayer;
       }
     );
   }
   this.showNew = false;
 }

 
 onEdit(index: number){
     this.selectedRow = index;
     this.instModel = new AppInst();
     this.instModel = Object.assign({}, this.appinsts[this.selectedRow]);
     this.submitType = 'Update';
     this.showNew = true;
 }

 onDelete(index: number) {
   this.selectedRow = index;
   this.instModel = new AppInst();
   this.instModel = Object.assign({}, this.appinsts[this.selectedRow]);
   this.data.deleteAppInst(this.instModel).subscribe(
     (data) => {
         // console.log("Delete called with index = ${index}");
         this.appinsts.splice(index,1);
     },
     error =>  {
       // Temporary solution: need to understand this error better.
       // console.log('oops', error);
       // console.log("Got Error on Delete  with index = " + index);
       this.appinsts.splice(index,1);
     }
   );
 }

 onCancel() {
   this.showNew = false;
 }


}
