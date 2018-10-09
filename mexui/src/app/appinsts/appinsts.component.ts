import { Component, OnInit } from '@angular/core';
import { DevdataService } from '../devdata.service';
import {AppInst, Flavor, App } from '../app.model';
import  {find, sortBy, trim}  from 'lodash';

class DeveloperItem {
  id: number     
  dvlpr: string  // Developer Name

  constructor(id, dvlpr){
      this.id = id;
      this.dvlpr = dvlpr;
  }
}

class AppItem {
  id: number
  dvlpr_id: number      // index of Developer Name in Developer list
  app: string        // Application Name

  constructor (id, dvlpr_id, app) {
  this.id = id;
  this.dvlpr_id = dvlpr_id;
  this.app = app;
  }
}

class VerItem {
  id: number
  dvlpr_id: number     // Developer Name
  app_id: number      // Application Name
  ver: string       // Version Name
  imagepath: string                   //        entry.result.image_path,
                     //        entry.result.image_type,
                    //        entry.result.access_layer,
                    //        entry.result.default_flavor.name,
                    //        entry.result.cluster.name

  constructor (id, dvlpr_id, app_id, ver) {
  this.id = id;
  this.dvlpr_id = dvlpr_id;
  this.app_id = app_id;
  this.ver = ver;
  }
}

class CloudletItem {
  id: number
  cld: string

  constructor(id, cld) {
    this.id = id;
    this.cld = cld;
  }
}


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

  flavors: string[] = [];
  developers: DeveloperItem[] = [];
  applications: AppItem[] = [];
  versions: VerItem[] = [];
  cloudlets: CloudletItem[] = [];

  selectedDvlpr: number = 0;
  selectedApp: number = 0;
  selectedVer: number = 0;
  selectedDvlprList:DeveloperItem[]  = [];
  selectedAppList: AppItem[] = [];
  selectedVerList: VerItem[] = [];
  selectedCld: number = 0;


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

    this.data.getApps().subscribe(
      (data) => {
        var obj  = JSON.parse("[" + data.split('}\n{').join('},\n{') + "]")
        var devlist: DeveloperItem[] = [];
        var devid = 1;

        var applist: AppItem[] = [];
        var appid = 1;

        var verlist: VerItem[] = [];
        var verid = 1;

        obj.forEach(function(entry){
          var index = -1; // -1 return value from array findIndex means no matching element was found
          var found_devid = 0;  // 0 is not a valid devid as devid starts from 1
          var found_appid = 0;  // 0 is not a valid appid as appid starts from 1
          var found_verid = 0;  // 0 is not a valid verid as verid starts from 1
          index = devlist.findIndex(function (el) {
            return el.dvlpr === entry.result.key.developer_key.name;
          });
          
          if (index === -1) {
            // Developer name was not found in its array, so we need to insert it and assign it a unique devid.
              found_devid = devid;
              devlist.push(new DeveloperItem(devid++, trim(entry.result.key.developer_key.name)));
          } else {
            found_devid = devlist[index].id;
          }
          // Reset index for search in applist
          index = -1;
          index = applist.findIndex(function (el) {
            return (el.dvlpr_id === found_devid &&
                    el.app === entry.result.key.name);
          });

          if (index === -1) {
            // App was not index in the applist, so need to insert
            found_appid = appid;
            applist.push(new AppItem( appid++, found_devid,
              trim(entry.result.key.name)
             ));
          } else {
            found_appid = applist[index].id;
          }
          // Reset index for version search
          index = -1;
          index = verlist.findIndex(function (el) {
            return (el.dvlpr_id === found_devid &&
                    el.app_id === found_appid &&
                    el.ver === entry.result.key.version
                  );
          });

          if (index === -1) {
            //Index is not in the verlist array
            verlist.push(new VerItem(verid++,
              found_devid,
              found_appid,
              trim(entry.result.key.version)
             ));
          } else {
            found_verid = verlist[index].id;
          }
          //  applist.push(new App(
          //        entry.result.key.developer_key.name,
          //        entry.result.key.name,
          //        entry.result.key.version,
          //        entry.result.image_path,
          //        entry.result.image_type,
          //        entry.result.access_layer,
          //        entry.result.default_flavor.name,
          //        entry.result.cluster.name
          //  )
          // )
        });
        this.developers = devlist;
        this.applications = applist;
        this.versions = verlist;
        // console.log(this.developers);
        // console.log(this.applications);
        // console.log(this.versions);
      }
   );

    this.data.getFlavors().subscribe(
      (data) => {
        var obj = JSON.parse("[" + data.split('}\n{').join('},\n{') + "]")
        var flvlist: string[] = [];
        obj.forEach(function(entry) {
          flvlist.push(entry.result.key.name);
      });
        this.flavors = flvlist;
        //console.log(this.flavors);
      }
   );

   this.data.getCloudlets().subscribe(
    (data) => {
      // console.log(data);
      var obj = JSON.parse( "[" + data.split('}\n{').join('},\n{') + "]");
      var cldid: number = 1;
      var cldlist: CloudletItem[] = [];
      obj.forEach(function(entry) {
        cldlist.push (new CloudletItem(
                 cldid++,
                 entry.result.key.operator_key.name + "---" + entry.result.key.name
        ));
        // cldlist.push(new Cloudlet(entry.result.key.operator_key.name, 
        //                                    entry.result.key.name,
        //                                    entry.result.access_uri,
        //                                    entry.result.location.lat,
        //                                    entry.result.location.long,
        //                                    entry.result.ip_support,
        //                                    entry.result.num_dynamic_ips
        //                                   ));
    });
      
      this.cloudlets = cldlist;
      // console.log(this.cloudlets);
    }
 );
  }

  onSelectCloudlet(cld_id: number){
    this.selectedCld = cld_id;
    console.log('selectedCld is ' + cld_id)
  }

  onSelectDeveloper(dvlpr_id: number){
       this.selectedDvlpr = +dvlpr_id;
      //  console.log('app is' + this.selectedApp );
      //  console.log('ver_id is' + this.selectedVer);
      //  console.log('dvlpr is' + this.selectedDvlpr );
       this.selectedApp = 0;
       this.selectedVer = 0;
        //Update the value of instModel.developerName based on the selected developer

       this.selectedAppList= this.applications.filter((item)=>{
             return item.dvlpr_id === +dvlpr_id
       })
       this.selectedVerList=[];
       this.selectedDvlprList = this.developers;
  }

  onSelectApp(app_id: number) {
    this.selectedApp = app_id;
    console.log('app is ' + this.selectedApp );
    console.log('ver_id is ' + this.selectedVer);
    console.log('dvlpr is ' + this.selectedDvlpr );
    

    //Update the value of instModel.name based on the selected application
    
    this.selectedVerList = this.versions.filter( (item) => {
        return (item.dvlpr_id === +this.selectedDvlpr && item.app_id === app_id)
    })
  }

  onSelectVer(ver_id: number) {
    this.selectedVer = ver_id;
    console.log('ver_id is ' + ver_id );
    console.log('dvlpr is ' + this.selectedDvlpr );
    console.log('app is ' + this.selectedApp );
    //Update the value of instModel.version based on the selected version
  }
  onNew(){
    this.instModel = new AppInst();
    // this.instModel.liveNess="LivenessStatic";
    // this.instModel.imageType="ImageTypeDocker";
    // this.instModel.accessLayer="AccessLayer7";
    this.instModel.liveNess="1";
    this.instModel.imageType="1";
    this.instModel.accessLayer="2";
    this.submitType = 'Save';
    this.showNew = true;
 }

 onSave(){
   
   // Update corresponding instModel field for Developer
   if (this.selectedDvlpr === 0) {
     alert('Please Select Developer');
     return;
   } else {
     // console.log('selectedDvlpr is ' + this.selectedDvlpr );
     // console.log(this.developers);
    
     for ( var index:number  = 0, len:number = this.developers.length; index < len ; index++ ){
      // console.log( 'inside for loop iteration index' + index);
      // console.log( 'len is' + len);
      // console.log('this.developers[index].id is' + this.developers[index].id);
      // console.log('this.developers[index].dvlpr is', this.developers[index].dvlpr);
      // console.log('this.selectedDvlpr is', this.selectedDvlpr);
       if (this.developers[index].id === +this.selectedDvlpr) {
         this.instModel.developerName = this.developers[index].dvlpr;
         break;
       }
     }
   }
   // console.log(' the developerName is ' + this.instModel.developerName);
   
   // Update corresponding instModel field for App
   if (this.selectedApp === 0) {
    alert('Please Select Application');
    return;
  } else {
    console.log('selectedApp is ' + this.selectedApp );
    console.log(this.applications);
   
    for ( var index:number  = 0, len:number = this.applications.length; index < len ; index++ ){
     console.log( 'inside for loop iteration index' + index);
     console.log( 'len is' + len);
     console.log('this.applications[index].id is' + this.applications[index].id);
     console.log('this.applications[index].app is', this.applications[index].app);
     console.log('this.selectedApp is', this.selectedApp);
      if (this.applications[index].id === +this.selectedApp  &&
          this.applications[index].dvlpr_id === +this.selectedDvlpr) {
        this.instModel.name = this.applications[index].app;
        break;
      }
    }
  }
  // console.log(' the applicationName is ' + this.instModel.name);

  // Update corresponding instModel field for Version
  if (this.selectedVer === 0) {
    alert('Please Select Version');
    return;
  } else {
    console.log('selectedVer is ' + this.selectedVer );
    console.log(this.versions);
   
    for ( var index:number  = 0, len:number = this.versions.length; index < len ; index++ ){
     console.log( 'inside for loop iteration index' + index);
     console.log( 'len is' + len);
     console.log('this.versions[index].id is' + this.versions[index].id);
     console.log('this.versions[index].ver is', this.versions[index].ver);
     console.log('this.selectedVer is', this.selectedVer);
      if (this.versions[index].id === +this.selectedVer &&
          this.versions[index].app_id === +this.selectedApp &&
          this.versions[index].dvlpr_id === +this.selectedDvlpr
      ) {
        this.instModel.version = this.versions[index].ver;
        break;
      }
    }
  }

  // Update corresponding instModel field for Cloudlet
  if (this.selectedCld === 0) {
    alert('Please Select Cloudlet');
    return;
  } else {
    console.log('selectedCld is ' + this.selectedCld );
    console.log(this.cloudlets);
   
    for ( var index:number  = 0, len:number = this.cloudlets.length; index < len ; index++ ){
     console.log( 'inside for loop iteration index' + index);
     console.log( 'len is' + len);
     console.log('this.cloudlets[index].id is' + this.cloudlets[index].id);
     console.log('this.cloudlets[index].cld is', this.cloudlets[index].cld);
     console.log('this.selectedCld is', this.selectedCld);
      if (this.cloudlets[index].id === +this.selectedCld) {
         
        [this.instModel.cloudletKey, this.instModel.cloudletName] = this.cloudlets[index].cld.split("---");
    
        break;
      }
    }
  }


   if (this.submitType === 'Save') {
     
     this.data.createAppInst(this.instModel).subscribe(
        (data) => {
         //this.appinsts.push(this.instModel);
         this.ngOnInit();
          // console.log(this.instModel);
        },
      (error) => {
        // console.log ("We should not get here!");
        console.log(this.instModel);
        //this.appinsts.push(this.instModel);
        this.ngOnInit();
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
         this.ngOnInit();
     },
     error =>  {
       // Temporary solution: need to understand this error better.
       // console.log('oops', error);
       // console.log("Got Error on Delete  with index = " + index);
       this.appinsts.splice(index,1);
       this.ngOnInit();
     }
   );
 }

 onCancel() {
   this.showNew = false;
 }

 onChangeFlavor(flavor: string) {
   this.instModel.flavor = flavor;
 }

 

}
