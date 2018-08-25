import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/Observable'; 
import 'rxjs/add/observable/of';
import { map, catchError } from 'rxjs/operators';


@Injectable({
  providedIn: 'root'
})
export class DevdataService {
  obj: Object;
  constructor(private http: HttpClient) { }

  

  getDevelopers() {
    return this.http.post('http://0.0.0.0:36002/show/developer',"",{responseType:'text'});
  }
  
  getApps(){
    return this.http.post('http://0.0.0.0:36002/show/app',"",{responseType:'text'});
  }
  getAppInstances() {
    return this.http.post('http://0.0.0.0:36002/show/appinst',"",{responseType:'text'});
  }
  

  getFlavors() {
    return this.http.post("http://0.0.0.0:36002/show/flavor","",{responseType:'text'});
  }

  getOperators() {
    return this.http.post('http://0.0.0.0:36002/show/operator',"",{responseType:'text'});
  }

  getCloudlets() {
    return this.http.post('http://0.0.0.0:36002/show/cloudlet',"",{responseType:'text'});
  }

  getClusters() {
    return this.http.post('http://0.0.0.0:36002/show/cluster',"",{responseType:'text'});
  }
}
 