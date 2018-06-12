# setup-local 

Starts applications for Controller, Etcd, CRM and DME on the local machine.  

Usage: 
setup-local [options]

options:
  -action string
        [start stop status cleanup]
  -datafile string
        optional yml data file
  -deployment string
        [process container]
  -processfile string
        mandatory yml application startup file


Sample Files:
- process.yml -- simple process setup with no redundancy except etcd
- process-multi.yml -- multiple instances of each process
- data.yml -- example provisioning

Logs for each process are created in the current directory.

Examples:
./setup-local  -action start -datafile ./data.yml  -processfile ./processes.yml -deployment process
./setup-local  -action stop -datafile ./data.yml  -processfile ./processes.yml -deployment process


