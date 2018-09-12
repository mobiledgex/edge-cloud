import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import {Oprator} from '../app.model';

@Component({
  selector: 'app-operators',
  templateUrl: './operators.component.html',
  styleUrls: ['./operators.component.scss']
})
export class OperatorsComponent implements OnInit {
  operators$: Object;
  operators: Oprator[] = [];
  oprModel: Oprator;
  showNew: Boolean = false;    
  submitType: string = 'Save';   
  selectedRow: number;           
  
  constructor(private data: DevdataService) { }

  ngOnInit() {
    this.data.getOperators().subscribe(
      (data) => {
        var obj =  JSON.parse("[" + data.split('}\n{').join('},\n{') + "]");
        var oprlist : Oprator[] = [];
        obj.forEach(function(entry) {
          oprlist.push(new Oprator(entry.result.key.name));
        });
        this.operators$ = obj;
        this.operators = oprlist;
      });   
  }
  
  onNew(){
    this.oprModel = new Oprator();
    this.submitType = 'Save';
    this.showNew = true;
 }

 onSave(){
   if (this.submitType === 'Save') {
     
     this.data.createOperator(this.oprModel).subscribe(
        (data) => {
         this.operators.push(this.oprModel);
          // console.log(this.oprModel);
        }
     );
   } else {
     this.data.updateOperator(this.oprModel).subscribe(
       (data) => {
           // Update existing properties based on the model
           this.operators[this.selectedRow].name = this.oprModel.name;
       }
     );
   }
   this.showNew = false;
 }

 
 onEdit(index: number){
     this.selectedRow = index;
     this.oprModel = new Oprator();
     this.oprModel = Object.assign({}, this.operators[this.selectedRow]);
     this.submitType = 'Update';
     this.showNew = true;
 }

 onDelete(index: number) {
   this.selectedRow = index;
   this.oprModel = new Oprator();
   this.oprModel = Object.assign({}, this.operators[this.selectedRow]);
   this.data.deleteOperator(this.oprModel).subscribe(
     (data) => {
         this.operators.splice(index,1);
     }
   );
 }

 onCancel() {
   this.showNew = false;
 }
}


