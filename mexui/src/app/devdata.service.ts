import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { HttpHeaders } from '@angular/common/http';
import { Observable, throwError, Operator } from 'rxjs';
import { catchError, retry } from 'rxjs/operators';
import 'rxjs/add/observable/of';
import { Developer, App, AppInst, Flavor, Oprator, Cloudlet, Cluster } from './app.model';

const ctrlUrl="https://mexdemo.ctrl.mobiledgex.net:36001";
//const ctrlUrl="http://mexdemo.ctrl-http.mobiledgex.net:36001";
// const ctrlUrl="http://0.0.0.0:36002";


@Injectable({
  providedIn: 'root'
})


export class DevdataService {
  obj: Object;

  

  constructor(private http: HttpClient) { }


  private handleError(error: HttpErrorResponse) {
    if (error.error instanceof ErrorEvent) {
    // A client-side or network error occurred. Handle it accordingly.
    console.error('An error occurred:', error.error.message);
  } else {
    // The backend returned an unsuccessful response code.
    // The response body may contain clues as to what went wrong,
    console.error(
      `Backend returned code ${error.status}, ` +
      `body was: ${error.error}`);
  }
  // return an observable with a user-facing error message
  return throwError(
    'Something bad happened; please try again later.');
};
  

  getDevelopers() {
    return this.http.post(`${ctrlUrl}/show/developer`,"",{responseType:'text'});
  }
  
  createDeveloper(dev: Developer){
    
    var body: string = `{
        "key": {
          "name": "${dev.developerName}"
        },
        "username": "${dev.userName}",
        "passhash": "${dev.passHash}",
        "address": "${dev.address}",
        "email": "${dev.email}"
      }`;
    
    return this.http.post(`$ctrlUrl/create/developer`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  updateDeveloper(dev: Developer){
    var body: string = `{
      "key": {
        "name": "${dev.developerName}"
      },
      "username": "${dev.userName}",
      "passhash": "${dev.passHash}",
      "address": "${dev.address}",
      "email": "${dev.email}"
    }`;
    console.log(body);
    return this.http.post(`${ctrlUrl}/update/developer`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  deleteDeveloper(dev:Developer) {
    var body: string = `{
      "key": {
        "name": "${dev.developerName}"
      },
      "username": "${dev.userName}",
      "passhash": "${dev.passHash}",
      "address": "${dev.address}",
      "email": "${dev.email}"
    }`;
    return this.http.post(`${ctrlUrl}/delete/developer`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }


  getApps(){
    return this.http.post(`${ctrlUrl}/show/app`,"",{responseType:'text'});
  }

  createApp(app: App){
    
    var body: string = `{
          "key": {
            "developer_key": {
              "name": "${app.developerName}"
            },
            "name": "${app.name}",
            "version": "${app.version}"
          },
          "image_path": "${app.imagePath}",
          "image_type": "${app.imageType}",
          "access_layer": "${app.accessLayer}",
          "default_flavor": {
            "name": "${app.defaultFlavor}"
          },
          "cluster": {
            "name": "${app.cluster}"
          }
        }`;
    
        console.log(body);
    return this.http.post(`${ctrlUrl}/create/app`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  updateApp(app: App){
    var body: string = `{
      "key": {
        "developer_key": {
          "name": "${app.developerName}"
        },
        "name": "${app.name}",
        "version": "${app.version}"
      },
      "image_path": "${app.imagePath}",
      "image_type": "${app.imageType}",
      "access_layer": "${app.accessLayer}",
      "default_flavor": {
        "name": "${app.defaultFlavor}"
      },
      "cluster": {
        "name": "${app.cluster}"
      }
    }`;

    console.log(body);
    return this.http.post(`${ctrlUrl}/update/app`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  deleteApp(app:App) {
    var body: string = `{
          "key": {
            "developer_key": {
              "name": "${app.developerName}"
            },
            "name": "${app.name}",
            "version": "${app.version}"
          },
          "image_path": "${app.imagePath}",
          "image_type": "${app.imageType}",
          "access_layer": "${app.accessLayer}",
          "default_flavor": {
            "name": "${app.defaultFlavor}"
          },
          "cluster": {
            "name": "${app.cluster}"
          }
        }`;
        console.log(body);
    return this.http.post(`${ctrlUrl}/delete/app`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  getAppInstances() {
    return this.http.post(`${ctrlUrl}/show/appinst`,"",{responseType:'text'});
  }
  
  createAppInst(inst: AppInst){
    
    var body: string = `{
      "key": {
        "app_key": {
          "developer_key": {
            "name": "${inst.developerName}"
          },
          "name": "${inst.name}",
          "version": "${inst.version}"
        },
        "cloudlet_key": {
          "operator_key": {
            "name": "${inst.cloudletKey}"
          },
          "name": "${inst.cloudletName}"
        },
        "id": ${inst.id}
      },
      "cloudlet_loc": {
        "lat": ${inst.cloudletLat},
        "long": ${inst.cloudletLong}
      },
      "uri": "${inst.uri}",
      "cluster_inst_key": {
        "cluster_key": {
          "name": "${inst.cluster}"
        },
        "cloudlet_key": {
          "operator_key": {
            "name": "${inst.cloudletKey}"
          },
          "name": "${inst.cloudletName}"
        }
      },
      "liveness": ${inst.liveNess},
      "image_path": "${inst.imagePath}",
      "image_type": ${inst.imageType},
      "mapped_path": "${inst.mappedPath}",
      "flavor": {
        "name": "${inst.flavor}"
      },
      "access_layer": ${inst.accessLayer}
    }`;
   
    // "mapped_ports": "${inst.mappedPorts}",  For some reason create request chokes on mapped_ports
        console.log(body);
    return this.http.post(`${ctrlUrl}/create/appinst`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  updateAppInst(inst: AppInst){
    var body: string = `{
      "key": {
        "app_key": {
          "developer_key": {
            "name": "${inst.developerName}"
          },
          "name": "${inst.name}",
          "version": "${inst.version}"
        },
        "cloudlet_key": {
          "operator_key": {
            "name": "${inst.cloudletKey}"
          },
          "name": "${inst.cloudletName}"
        },
        "id": ${inst.id}
      },
      "cloudlet_loc": {
        "lat": ${inst.cloudletLat},
        "long": ${inst.cloudletLong}
      },
      "uri": "${inst.uri}",
      "cluster_inst_key": {
        "cluster_key": {
          "name": "${inst.cluster}"
        },
        "cloudlet_key": {
          "operator_key": {
            "name": "${inst.cloudletKey}"
          },
          "name": "${inst.cloudletName}"
        }
      },
      "liveness": "${inst.liveNess}",
      "image_path": "${inst.imagePath}",
      "image_type": "${inst.imageType}",
      "mapped_path": "${inst.mappedPath}",
      "flavor": {
        "name": "${inst.flavor}"
      },
      "access_layer": "${inst.accessLayer}"
    }`;

    console.log(body);
    return this.http.post(`${ctrlUrl}/update/appinst`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  deleteAppInst(inst:AppInst) {
    var body: string = `{
      "key": {
        "app_key": {
          "developer_key": {
            "name": "${inst.developerName}"
          },
          "name": "${inst.name}",
          "version": "${inst.version}"
        },
        "cloudlet_key": {
          "operator_key": {
            "name": "${inst.cloudletKey}"
          },
          "name": "${inst.cloudletName}"
        },
        "id": ${inst.id}
      },
      "cloudlet_loc": {
        "lat": ${inst.cloudletLat},
        "long": ${inst.cloudletLong}
      },
      "uri": "${inst.uri}",
      "cluster_inst_key": {
        "cluster_key": {
          "name": "${inst.cluster}"
        },
        "cloudlet_key": {
          "operator_key": {
            "name": "${inst.cloudletKey}"
          },
          "name": "${inst.cloudletName}"
        }
      },
      "liveness": "${inst.liveNess}",
      "image_path": "${inst.imagePath}",
      "image_type": "${inst.imageType}",
      "mapped_path": "${inst.mappedPath}",
      "flavor": {
        "name": "${inst.flavor}"
      },
      "access_layer": "${inst.accessLayer}"
    }`;
        console.log(body);
    return this.http.post(`${ctrlUrl}/delete/appinst`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  getFlavors() {
    return this.http.post(`${ctrlUrl}/show/flavor`,"",{responseType:'text'});
  }

  createFlavor(flv: Flavor){
    
    var body: string = `{
      "key": {
        "name": "${flv.name}"
      },
      "ram": ${flv.ram},
      "vcpus": ${flv.vcpus},
      "disk": ${flv.disk}
    }`;
    
    return this.http.post(`${ctrlUrl}/create/flavor`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  updateFlavor(flv: Flavor){
    var body: string = `{
      "key": {
        "name": "${flv.name}"
      },
      "ram": ${flv.ram},
      "vcpus": ${flv.vcpus},
      "disk": ${flv.disk}
    }`;
    console.log(body);
    return this.http.post(`${ctrlUrl}/update/flavor`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  deleteFlavor(flv: Flavor) {
    var body: string = `{
      "key": {
        "name": "${flv.name}"
      },
      "ram": ${flv.ram},
      "vcpus": ${flv.vcpus},
      "disk": ${flv.disk}
    }`;
    return this.http.post(`${ctrlUrl}/delete/flavor`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }


  getOperators() {
    return this.http.post(`${ctrlUrl}/show/operator`,"",{responseType:'text'});
  }
  createOperator(opr: Oprator){
    
    var body: string = `{
        "key": {
          "name": "${opr.name}"
        }
      }`;
    
    return this.http.post(`${ctrlUrl}/create/operator`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  updateOperator(opr: Oprator){
    var body: string = `{
      "key": {
        "name": "${opr.name}"
      }
    }`;
    console.log(body);
    return this.http.post(`${ctrlUrl}/update/operator`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  deleteOperator(opr: Oprator) {
    var body: string = `{
      "key": {
        "name": "${opr.name}"
      }
    }`;
    return this.http.post(`${ctrlUrl}/delete/operator`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }
  getCloudlets() {
    return this.http.post(`${ctrlUrl}/show/cloudlet`,"",{responseType:'text'});
  }
  createCloudlet(cld: Cloudlet){
    
    var body: string = `{
                    "key": {
                      "operator_key": {
                        "name": "${cld.operatorName}"
                      },
                      "name": "${cld.cloudletName}"
                    },
                    "access_uri": "${cld.uri}",
                    "location": {
                      "lat": ${cld.locationLat},
                      "long": ${cld.locationLong}
                    },
                    "ip_support": "${cld.ipSupport}",
                    "num_dynamic_ips": ${cld.numDynamicIps}
                  }`;
    
                  console.log(body);
                  console.log(`${ctrlUrl}/create/cloudlet`);
    return this.http.post(`${ctrlUrl}/create/cloudlet`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  updateCloudlet(cld: Cloudlet){
    var body: string = `{
      "key": {
        "operator_key": {
          "name": "${cld.operatorName}"
        },
        "name": "${cld.cloudletName}"
      },
      "access_uri": "${cld.uri}",
      "location": {
        "lat": ${cld.locationLat},
        "long": ${cld.locationLong}
      },
      "ip_support": "${cld.ipSupport}",
      "num_dynamic_ips": ${cld.numDynamicIps}
    }`;
    console.log(body);
    return this.http.post(`${ctrlUrl}/update/cloudlet`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  deleteCloudlet(cld: Cloudlet){
    var body: string = `{
      "key": {
        "operator_key": {
          "name": "${cld.operatorName}"
        },
        "name": "${cld.cloudletName}"
      },
      "access_uri": "${cld.uri}",
      "location": {
        "lat": ${cld.locationLat},
        "long": ${cld.locationLong}
      },
      "ip_support": "${cld.ipSupport}",
      "num_dynamic_ips": ${cld.numDynamicIps}
    }`;
    console.log(body);
    console.log(`${ctrlUrl}/delete/cloudet`);
    return this.http.post(`${ctrlUrl}/delete/cloudet`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }
  getClusters() {
    return this.http.post(`${ctrlUrl}/show/cluster`,"",{responseType:'text'});
  }
  createCluster(cls : Cluster) {
    
    var body: string = `{
        "key": {
          "name": "${cls.name}"
        },
        "default_flavor": {
          "name": "${cls.defaultFlavor}"
        }
      }`;
    
    return this.http.post(`${ctrlUrl}/create/cluster`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  updateCluster(cls : Cluster) {
    var body: string = `{
      "key": {
        "name": "${cls.name}"
      },
      "default_flavor": {
        "name": "${cls.defaultFlavor}"
      }
    }`;
    console.log(body);
    return this.http.post(`${ctrlUrl}/update/cluster`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  deleteCluster(cls : Cluster) {
    var body: string = `{
      "key": {
        "name": "${cls.name}"
      },
      "default_flavor": {
        "name": "${cls.defaultFlavor}"
      }
    }`;
    return this.http.post(`${ctrlUrl}/delete/cluster`, body, {})
    .pipe(
      catchError(this.handleError)
    );
  }

  getNodes() {
    return this.http.post(`${ctrlUrl}/show/node`,"",{responseType:'text'});
  }
}

