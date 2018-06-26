  
Deploys, starts, stops, provisions services on either the local MAC or remote machines.  Any combo thereof.


Prior to remote deployment, a "make linux" should be performed at the top level of the repo.

Usage:
setup-mex [options]

options:
  -actions string
        one or more of: [start stop delete cleanup fetchlogs status show update create deploy] separated by ,
  -datafile string
setup-mex [options]

options:
  -actions string
        one or more of: [start stop delete cleanup fetchlogs status show update create deploy] separated by ,
  -datafile string
        optional yml data file
  -deployment string
setup-local -setupfile three_vm/setup.yml -action stop,fetchlogs,cleanup -outputdir /tmp/setupmex
  -datafile string
        optional yml data file
  -deployment string
        [process container] (default "process")
  -outputdir string
        option directory to store output and logs, TS suffix will be replaced with timestamp
  -setupfile string
        mandatory yml topology file
  -timestamp
        append current timestamp to outputdir

Sample Setups are in the setups directory:

local_multi -- local processes with multiple controllers
local_simplex -- local processes with one instance of each process (except etcd)
three_vm -- VBOX deployment against 3 different VMs
gcp_2vm -- 2 VM deployment in google cloud

Logs for each process are created in the current directory.

Examples:
setup-mex -setupfile three_vm/setup.yml -action deploy
setup-mex -setupfile three_vm/setup.yml -action start -datafile three_vm/data.yml
setup-mex -setupfile three_vm/setup.yml -action create -datafile three_vm/data.yml
setup-mex -setupfile three_vm/setup.yml -action update -datafile three_vm/data.yml
setup-mex -setupfile three_vm/setup.yml -action stop,fetchlogs,cleanup -outputdir /tmp/setupmex
