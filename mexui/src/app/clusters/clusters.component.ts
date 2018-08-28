import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import { Cluster } from '../app.model';

@Component({
  selector: 'app-clusters',
  templateUrl: './clusters.component.html',
  styleUrls: ['./clusters.component.scss']
})
export class ClustersComponent implements OnInit {
  clusters$: Object;
  clusters: Cluster[] = [];  
  clsModel: Cluster;           
  showNew: Boolean = false;    
  submitType: string = 'Save';   
  selectedRow: number;   
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getClusters().subscribe(
      (data) => {
        var obj =  JSON.parse("[" + data.split('}\n{').join('},\n{') + "]");
        var clslist: Cluster[] = [];
        obj.forEach(function(entry) {
          clslist.push(new Cluster(entry.result.key.name, 
                                             entry.result.default_flavor.name
                                            ));
         });

        this.clusters$ = obj;
        this.clusters = clslist;
      }
   );
  }
  onNew(){
    this.clsModel = new Cluster();
    this.submitType = 'Save';
    this.showNew = true;
 }

 onSave(){
   if (this.submitType === 'Save') {
     
     this.data.createCluster(this.clsModel).subscribe(
        (data) => {
         this.clusters.push(this.clsModel);
          // console.log(this.clsModel);
        }
     );
   } else {
     this.data.updateCluster(this.clsModel).subscribe(
       (data) => {
     // Update existing properties based on the model
     this.clusters[this.selectedRow].name = this.clsModel.name;
     this.clusters[this.selectedRow].defaultFlavor = this.clsModel.defaultFlavor;
       }
     );
   }
   this.showNew = false;
 }

 
 onEdit(index: number){
     this.selectedRow = index;
     this.clsModel = new Cluster();
     this.clsModel = Object.assign({}, this.clusters[this.selectedRow]);
     this.submitType = 'Update';
     this.showNew = true;
 }

 onDelete(index: number) {
   this.selectedRow = index;
   this.clsModel = new Cluster();
   this.clsModel = Object.assign({}, this.clusters[this.selectedRow]);
   this.data.deleteCluster(this.clsModel).subscribe(
     (data) => {
         this.clusters.splice(index,1);
     }
   );
 }

 onCancel() {
   this.showNew = false;
 }

}
