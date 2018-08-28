import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import {App} from '../app.model';


@Component({
  selector: 'app-apps',
  templateUrl: './apps.component.html',
  styleUrls: ['./apps.component.scss']
})
export class AppsComponent implements OnInit {
  apps$: Object;

  apps: App[] = [];
  appModel: App;
  showNew: Boolean = false;
  submitType: string = 'Save';
  selectedRow: number;

  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getApps().subscribe(
      (data) => {
        var obj  = JSON.parse("[" + data.split('}\n{').join('},\n{') + "]")
        var applist: App[] = [];
        obj.forEach(function(entry){
           applist.push(new App(
                 entry.result.key.developer_key.name,
                 entry.result.key.name,
                 entry.result.key.version,
                 entry.result.image_path,
                 entry.result.image_type,
                 entry.result.access_layer,
                 entry.result.default_flavor.name,
                 entry.result.cluster.name
           )
          )
        });
        this.apps$ = obj;
        this.apps = applist;
      }
   );
  }
 
  onNew(){
    this.appModel = new App();
    this.submitType = 'Save';
    this.showNew = true;
 }

 onSave(){
   if (this.submitType === 'Save') {
     
     this.data.createApp(this.appModel).subscribe(
        (data) => {
         this.apps.push(this.appModel);
          // console.log(this.appModel);
        }
     );
   } else {
     this.data.updateApp(this.appModel).subscribe(
       (data) => {
     // Update existing properties based on the model
     this.apps[this.selectedRow].developerName = this.appModel.developerName;
     this.apps[this.selectedRow].name = this.appModel.name;
     this.apps[this.selectedRow].version = this.appModel.version;
     this.apps[this.selectedRow].imagePath  = this.appModel.imagePath;
     this.apps[this.selectedRow].imageType = this.appModel.imageType;
     this.apps[this.selectedRow].accessLayer = this.appModel.accessLayer;
     this.apps[this.selectedRow].defaultFlavor = this.appModel.defaultFlavor;
     this.apps[this.selectedRow].cluster = this.appModel.cluster;
       }
     );
   }
   this.showNew = false;
 }

 
 onEdit(index: number){
     this.selectedRow = index;
     this.appModel = new App();
     this.appModel = Object.assign({}, this.apps[this.selectedRow]);
     this.submitType = 'Update';
     this.showNew = true;
 }

 onDelete(index: number) {
   this.selectedRow = index;
   this.appModel = new App();
   this.appModel = Object.assign({}, this.apps[this.selectedRow]);
   this.data.deleteApp(this.appModel).subscribe(
     (data) => {
         this.apps.splice(index,1);
     }
   );
 }

 onCancel() {
   this.showNew = false;
 }
}
