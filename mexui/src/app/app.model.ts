export  class Developer {
    constructor(
       public developerName: string = '',
       public userName: string = '',
       public passHash: string = '',
       public address: string = '',
       public email: string = ''
    ){}
  }

// Apps fields
  // <li> Developer Name          : {{ ap.result.key.developer_key.name }}</li>
  // <li> Application Name        : {{ ap.result.key.name }}</li>
  // <li> Version                 : {{ ap.result.key.version }}</li>
  // <li> ImagePath               : {{ ap.result.image_path }}</li>
  // <li> ImageType               : {{ ap.result.image_type }}</li>
  // <li> AccessLayer             : {{ ap.result.access_layer }}</li>
  // <li> DefaultFlavor           : {{ ap.result.default_flavor.name }}</li>
  // <li> Cluster                 : {{ ap.result.cluster.name }} </li>

export class App {
  constructor(
    public developerName: string = '',
    public name: string = '',
    public version: string = '',
    public imagePath: string = '',
    public imageType: string = '',
    public accessLayer: string='',
    public defaultFlavor: string='',
    public cluster: string=''
  ){}
}

//AppInstance fields
            // <li> Id                     :  {{ appinst.result.key.id }}</li>
            // <li> Developer Name          : {{ appinst.result.key.app_key.developer_key.name }}</li>
            // <li> Application Name        : {{ appinst.result.key.app_key.name }}</li>
            // <li> Version                 : {{ appinst.result.key.app_key.version }}</li>
            // <li> ImagePath               : {{ appinst.result.image_path }}</li>
            // <li> ImageType               : {{ appinst.result.image_type }}</li>
            // <li> Uri                     : {{ appinst.result.uri }}</li>
            // <li> Liveness                : {{ appinst.result.liveness }}</li>
            // <li> Mapped Ports            : {{ appinst.result.mapped_ports }}</li>
            // <li> Mapped Path             : {{ appinst.result.mapped_path }}</li>
            // <li> Flavor                  : {{ appinst.result.flavor.name }}</li>
            // <li> Cloudlet Key            : {{ appinst.result.key.cloudlet_key.operator_key.name }}</li>
            // <li> Cloudlet Name            : {{ appinst.result.key.cloudlet_key.name }}</li>
            // <li> Cloudlet Lat            : {{ appinst.result.cloudlet_loc.lat }}</li>
            // <li> Cloudlet Long            : {{ appinst.result.cloudlet_loc.long }}</li>
            // <li> Cluster                 : {{ appinst.result.cluster_inst_key.cluster_key.name }} </li>

export class AppInst {
  constructor(
    public id: number = 0,
    public developerName: string = '',
    public name: string = '',
    public version: string = '',
    public imagePath: string = '',
    public imageType: string = '',
    public uri: string = '',
    public liveNess: string = '',
    public mappedPorts: string = '',
    public mappedPath: string = '',
    public flavor: string = '',
    public cloudletKey: string = '',
    public cloudletName: string = '',
    public cloudletLat: number = 0,
    public cloudletLong: number = 0,
    public cluster: string = '',
    public accessLayer: string='',
  ){}
}

// Flavor fields

            // <li> flavor Name: {{ flavor.result.key.name }}</li>
            // <li> ram    : {{ flavor.result.ram }}</li>
            // <li> vcpus  : {{ flavor.result.vcpus }}</li>
            // <li> disk    : {{ flavor.result.disk }}</li>
            // {
            //   "key": {
            //     "name": "x1.small"
            //   },
            //   "ram": 2048,
            //   "vcpus": 2,
            //   "disk": 2
            // }

 export class Flavor {
   constructor(
     public name  :string = 'x1.small',
     public ram   :number = 2048,
     public vcpus :number = 2,
     public disk  :number = 2
   ){}
 }           

// Operator Fields
// <li> Name: {{ operator.result.key.name }}</li>
//  {
//   "key": {
//     "name": "TMUS"
//   }
// }
export class Oprator {
  constructor(
    public name :string = 'TMUS'
  ){}
}

// Cloudlets fields
//             <li> Operator Name: {{ cloudlet.result.key.operator_key.name }}</li>
//             <li> Cloudlet Name    : {{ cloudlet.result.key.name }}</li>
//             <li> Access URI    : {{ cloudlet.result.access_uri }}</li>
//             <li> Longitude    : {{ cloudlet.result.location.long }}</li>
//             <li> Lattitude    : {{ cloudlet.result.location.lat }}</li>
//             <li> IP Support       : {{ cloudlet.result.ip_support }}</li>
//             <li> Number Of Dynamic IPs       : {{ cloudlet.result.num_dynamic_ips }}</li>
//
//             {
//               "key": {
//                 "operator_key": {
//                   "name": "TMUS"
//                 },
//                 "name": "tmocloud-1"
//               },
//               "access_uri": "cloud1.tmo",
//               "location": {
//                 "lat": 31,
//                 "long": -91
//               },
//               "ip_support": 2,
//               "num_dynamic_ips": 254
//             }
export class Cloudlet {
  constructor(
    public operatorName :string = '',
    public cloudletName :string = '',
    public uri :string = '',
    public locationLat :number = 0,
    public locationLong :number = 0,
    public ipSupport :string = 'IpSupportDynamic',
    public numDynamicIps :number = 0
  ){}
}

// Cluster Fields
//
// <li> Cluster Name: {{ cluster.result.key.name }}</li>
// <li> Cluster Default Flavor    : {{ cluster.result.default_flavor.name }}</li>

// {
//   "key": {
//     "name": "SmallCluster"
//   },
//   "default_flavor": {
//     "name": "c1.small"
//   }
// }

export class Cluster {
  constructor(
    public name :string = '',
    public defaultFlavor :string = 'c1.small'
  ){}
}