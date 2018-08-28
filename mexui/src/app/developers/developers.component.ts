import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import {Developer} from '../app.model';


@Component({
  selector: 'app-developers',
  templateUrl: './developers.component.html',
  styleUrls: ['./developers.component.scss']
})
export class DevelopersComponent implements OnInit {
  developers$: Object;

  developers: Developer[] = [];  // Maintain list of developers
  devModel: Developer;           // Maintain Developer Model
  showNew: Boolean = false;      // Maintain Developer form display status. By default, it will be false.
  submitType: string = 'Save';   // Either Save or Update based on whether the operation is  New or Edit respectively.
  selectedRow: number;           // Table row index of the selected item


  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getDevelopers().subscribe(
       (data) => {
         // console.log(data);
         data = "[" + data.split("}\n{").join("},\n{") + "]";
        var obj = JSON.parse(data);
        var devlist: Developer[] = [];
        obj.forEach(function(entry) {
            devlist.push(new Developer(entry.result.key.name, 
                                               entry.result.username,
                                               entry.result.passhash,
                                               entry.result.address,
                                               entry.result.email
                                              ));
        });
        this.developers$ = obj;
        //console.log(devlist);
        this.developers = devlist;
       }
    );
  }

 
  onNew(){
     this.devModel = new Developer();
     this.submitType = 'Save';
     this.showNew = true;
  }

  onSave(){
    if (this.submitType === 'Save') {
      
      this.data.createDeveloper(this.devModel).subscribe(
         (data) => {
          this.developers.push(this.devModel);
           // console.log(this.devModel);
         }
      );
    } else {
      this.data.updateDeveloper(this.devModel).subscribe(
        (data) => {
        
          
     
      // Update existing properties based on the model
      this.developers[this.selectedRow].developerName = this.devModel.developerName;
      this.developers[this.selectedRow].userName = this.devModel.userName;
      this.developers[this.selectedRow].passHash = this.devModel.passHash;
      this.developers[this.selectedRow].address  = this.devModel.address;
      this.developers[this.selectedRow].email = this.devModel.email;
        }
      );
    }
    this.showNew = false;
  }

  
  onEdit(index: number){
      this.selectedRow = index;
      this.devModel = new Developer();
      this.devModel = Object.assign({}, this.developers[this.selectedRow]);
      this.submitType = 'Update';
      this.showNew = true;
  }

  onDelete(index: number) {
    this.selectedRow = index;
    this.devModel = new Developer();
    this.devModel = Object.assign({}, this.developers[this.selectedRow]);
    this.data.deleteDeveloper(this.devModel).subscribe(
      (data) => {
          this.developers.splice(index,1);
      }
    );
  }

  onCancel() {
    this.showNew = false;
  }
}
