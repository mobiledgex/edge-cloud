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

  getDevelopers():Observable<any[]>{
    var developers_data = [
      {
        "key": {
          "name": "Ever.ai"
        },
        "address": "1 Letterman Drive Building C, San Francisco, CA 94129",
        "email": "edge.everai.com"
      },
      {
        "key": {
          "name": "1000 realities"
        },
        "address": "Kamienna 43, 31-403 Krakow, Poland",
        "email": "edge.1000realities.com"
      },
      {
        "key": {
          "name": "Sierraware LLC"
        },
        "address": "1250 Oakmead Parkway Suite 210, Sunnyvalue, CA 94085",
        "email": "support@sierraware.com"
      },
      {
        "key": {
          "name": "AcmeAppCo"
        },
        "username": "acmeapp",
        "passhash": "8136f09c17354891c642b9b9f1722c34",
        "address": "123 Maple Street, Gainesville, FL 32604",
        "email": "acmeapp@xxxx.com"
      }
    ];
    
    return Observable.of(developers_data);
     // [{"key":{"name":"Ever.ai"},"address":"1 Letterman Drive Building C, San Francisco, CA 94129","email":"edge.everai.com"},{"key":{"name":"1000 realities"},"address":"Kamienna 43, 31-403 Krakow, Poland","email":"edge.1000realities.com"},{"key":{"name":"Sierraware LLC"},"address":"1250 Oakmead Parkway Suite 210, Sunnyvalue, CA 94085","email":"support@sierraware.com"},{"key":{"name":"AcmeAppCo"},"username":"acmeapp","passhash":"8136f09c17354891c642b9b9f1722c34","address":"123 Maple Street, Gainesville, FL 32604","email":"acmeapp@xxxx.com"}]);
    // return this.http.get('https://jsonplaceholder.typicode.com/users')
    //    return  this.http.post('http://0.0.0.0:36002/show/developer',"");
    // console.log(obj);
    // return obj;

    
  }

  getApps(){
    var apps_data = [
      {
        "key": {
          "developer_key": {
            "name": "AcmeAppCo"
          },
          "name": "someApplication",
          "version": "1.0"
        },
        "image_path": "mobiledgex_AcmeAppCo/someApplication:1.0",
        "image_type": 1,
        "access_layer": 2,
        "default_flavor": {
          "name": "x1.small"
        },
        "cluster": {
          "name": "SmallCluster"
        }
      }
    ];
    
    
      //[{"key":{"developer_key":{"name":"AcmeAppCo"},"name":"someApplication","version":"1.0"},"image_path":"mobiledgex_AcmeAppCo/someApplication:1.0","image_type":1,"access_layer":2,"default_flavor":{"name":"x1.small"},"cluster":{"name":"SmallCluster"}}] 
    //)
    // this.http.get(url).map(res:Response) => res.json().data);
    
    // console.log(temp);
    // return temp;
    return Observable.of(apps_data);
    // return this.http.post('http://0.0.0.0:36002/show/app',"",{observe: 'response', responseType: "json"})//;
    // return this.http.post('http://0.0.0.0:36002/show/app',"");
  }

  getAppInstances():Observable<any[]> {
    var appinst_data = [
      {
        "key": {
          "app_key": {
            "developer_key": {
              "name": "AcmeAppCo"
            },
            "name": "someApplication",
            "version": "1.0"
          },
          "cloudlet_key": {
            "operator_key": {
              "name": "TMUS"
            },
            "name": "tmocloud1"
          },
          "id": 123
        },
        "cloudlet_loc": {
          "lat": 31,
          "long": -91
        },
        "uri": "someApplication.tmocloud1.mobiledgex.net",
        "cluster_inst_key": {
          "cluster_key": {
            "name": "SmallCluster"
          },
          "cloudlet_key": {
            "operator_key": {
              "name": "TMUS"
            },
            "name": "tmocloud1"
          }
        },
        "liveness": 1,
        "image_path": "mobiledgex_AcmeAppCo/someApplication:1.0",
        "image_type": 1,
        "mapped_ports": null,
        "mapped_path": "someApplication",
        "flavor": {
          "name": "x1.small"
        }
      },
      {
        "key": {
          "app_key": {
            "developer_key": {
              "name": "AcmeAppCo"
            },
            "name": "someApplication",
            "version": "1.0"
          },
          "cloudlet_key": {
            "operator_key": {
              "name": "TMUS"
            },
            "name": "tmocloud2"
          },
          "id": 123
        },
        "cloudlet_loc": {
          "lat": 35,
          "long": -95
        },
        "uri": "someApplication.tmocloud2.mobiledgex.net",
        "cluster_inst_key": {
          "cluster_key": {
            "name": "SmallCluster"
          },
          "cloudlet_key": {
            "operator_key": {
              "name": "TMUS"
            },
            "name": "tmocloud2"
          }
        },
        "liveness": 1,
        "image_path": "mobiledgex_AcmeAppCo/someApplication:1.0",
        "image_type": 1,
        "mapped_ports": null,
        "mapped_path": "someApplication",
        "flavor": {
          "name": "x1.small"
        }
      }
    ];
    return Observable.of(appinst_data);
  }

  getFlavors() {
    var flavor_data = [
      {
        "result": {
         "fields": [],
         "key": {
          "name": "x1.medium"
         },
         "ram": "4096",
         "vcpus": "4",
         "disk": "4"
        }
       },
       {
        "result": {
         "fields": [],
         "key": {
          "name": "x1.tiny"
         },
         "ram": "1024",
         "vcpus": "1",
         "disk": "1"
        }
       },
       {
        "result": {
         "fields": [],
         "key": {
          "name": "x1.small"
         },
         "ram": "2048",
         "vcpus": "2",
         "disk": "2"
        }
       }
    ];
    return Observable.of(flavor_data);
    // return this.http.post("http://0.0.0.0:36002/show/flavor","");
  }

  getOperators() {
    var operators_data = [
      {
        "key": {
          "name": "ATT"
        }
      },
      {
        "key": {
          "name": "TMUS"
        }
      }
    ]

    return Observable.of(operators_data);
  }

  getCloudlets() {
    var cloudlets_data = [
      {
        "key": {
          "operator_key": {
            "name": "TMUS"
          },
          "name": "tmocloud-1"
        },
        "access_uri": "cloud1.tmo",
        "location": {
          "lat": 31,
          "long": -91
        },
        "ip_support": 2,
        "num_dynamic_ips": 254
      },
      {
        "key": {
          "operator_key": {
            "name": "TMUS"
          },
          "name": "tmocloud-2"
        },
        "access_uri": "cloud2.tmo",
        "location": {
          "lat": 35,
          "long": -95
        },
        "ip_support": 2,
        "num_dynamic_ips": 254
      }
    ];
    return Observable.of(cloudlets_data);
  }

  getClusters() {
    var clusters_data = [
      {
        "key": {
          "name": "SmallCluster"
        },
        "default_flavor": {
          "name": "c1.small"
        }
      }
    ];

    return Observable.of(clusters_data);
  }
}
 