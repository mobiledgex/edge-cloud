LB=10.95.84.63
SERVERS="10.95.84.69 10.95.84.74 10.95.84.75"
PORTS="80 443 50051 55001"

for port in $PORTS
do
    echo "adding service to $port"
    ipvsadm -A -t $LB:$port
      
    for svr in $SERVERS 
    do
       echo "  adding server $svr"
       ipvsadm -a -t $LB:$port -r $svr:$port -m
    done
done

