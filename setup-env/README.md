#setup-env

Deploys, starts, stops, provisions services on either the local MAC or remote machines.  Any combo thereof.

Remote deployment requires ansible to be installed (brew install ansible). And the ANSIBLE_DIR must be set, which should be the full path to the
Ansible directory within this package

Prior to remote deployment, a "make linux" should be performed at the top level of the repo.

All operations are performed via the setup-local tool (which may at some point be renamed).

Usage: 
setup-local [options]

Usage: 
setup-local [options]

options:
  -action string
        [cleanup start stop status update create deploy]
  -datafile string
        optional yml data file
  -deployment string
        [process container] (default "process")
  -setupfile string
        mandatory yml topology file


Sample Setups are in the setups directory:

local_multi -- local processes with multiple controllers
local_simplex -- local processes with one instance of each process (except etcd)
three_vm -- deployment against 3 differnet VMs

Logs for each process are created in the current directory.

Examples:
setup-local -setupfile three_vm/setup.yml -action deploy
setup-local -setupfile three_vm/setup.yml -action start -datafile three_vm/data.yml
setup-local -setupfile three_vm/setup.yml -action create -datafile three_vm/data.yml
setup-local -setupfile three_vm/setup.yml -action update -datafile three_vm/data.yml
setup-local -setupfile three_vm/setup.yml -action stop 
setup-local -setupfile three_vm/setup.yml -action cleanup 



