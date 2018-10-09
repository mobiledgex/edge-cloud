#!/usr/local/bin/python3

# copy needed proto files into 1 directory
# I originally tried building these from the original path but ran into problems

import os
import shutil
import glob
import subprocess

edgecloud_dir = '/Users/andyanderson/go/src/github.com/mobiledgex/edge-cloud/'

protos_src_list = (edgecloud_dir + 'vendor/github.com/gogo/googleapis/google/api/',
                   edgecloud_dir + 'vendor/github.com/gogo/protobuf/gogoproto/',
                   #edgecloud_dir + 'vendor/github.com/golang/protobuf/protoc-gen-go/descriptor/',
                   edgecloud_dir + 'vendor/github.com/golang/protobuf/ptypes/timestamp/',
                   edgecloud_dir + 'd-match-engine/dme-proto/',
                   edgecloud_dir + 'edgeproto/',
                   edgecloud_dir + 'protoc-gen-cmd/protocmd/',
                   edgecloud_dir + 'protogen/',
)
protos_dest = edgecloud_dir + 'setup-env/e2e-tests/python/protos'

#file_skip_convert_list = ('annotations.proto')
import_proto_skip_convert_list = ('descriptor.proto')  # dont remove path from this since it causes a warning to print

generate_proto_cmd = 'python3 -m grpc_tools.protoc -I{} --python_out={} --grpc_python_out={} '.format(protos_dest, protos_dest, protos_dest)

# copy protofiles and change import statement to remove path info
for proto in protos_src_list:
    #flist = os.listdir(proto)
    flist = glob.glob(proto + '*.proto')
    #print(flist)
    for file in flist:
        print('copy {} to {}'.format(file, protos_dest))
        filename = os.path.basename(file)
        dest_file_copy = protos_dest + '/' + filename + '.orig'
        dest_file_new = protos_dest + '/' + filename
        shutil.copy2(file, dest_file_copy)

        with open(dest_file_copy) as read_file, open(dest_file_new, 'w') as write_file:
            line = read_file.readline()
            while line:
                if line.startswith('import') and '/' in line: # remove path from import statement
                    #print('found',line)
                    lsplit = line.split('"')
                    if os.path.basename(lsplit[1]) not in import_proto_skip_convert_list:
                        line = lsplit[0] + '"' + os.path.basename(lsplit[1]) + '"' + lsplit[2]
                    else:
                        print('skipping convert of', os.path.basename(lsplit[1]))    
                write_file.write(line)
                line = read_file.readline()

# convert proto files to python
flist_proto = glob.glob(protos_dest + '/*.proto')
for file in flist_proto:
    new_cmd = generate_proto_cmd + '/' + file
    print('executing', new_cmd)
    return_code = subprocess.call(new_cmd, shell=True)
    if return_code != 0:
        print('error converting proto. returncode=', return_code)

            
                   
