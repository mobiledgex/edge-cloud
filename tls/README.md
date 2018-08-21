TLS certs can be generated here with certstrap.

Both client and server certs must be signed by the private Certificate Authority mex-ca.crt

Generating CA File:
$ certstrap  init --common-name mex-ca
Enter passphrase (empty for no passphrase): 
<leave blank>
Enter same passphrase again: 
<leave blank>
Created out/mex-ca.key
Created out/mex-ca.crt
Created out/mex-ca.crl

The CA file can be re-used for every deployment.  It is included for every server and client.

Generating Server Certs:

Domain based, can be wildcard or FQDN:
$ certstrap request-cert --domain dme.xyz.mobiledgex.net
Enter passphrase (empty for no passphrase): 
<leave blank>
Enter same passphrase again: 
<leave blank>

Created out/dme.xyz.mobiledgex.net.key
Created out/dme.xyz.mobiledgex.net.csr

$ certstrap sign --CA mex-ca dme.xyz.mobiledgex.net
$ certstrap sign --CA mex-ca dme.xyx.mobiledgex.net

Created out/dme.xyz.mobiledgex.net.crt from out/dme.xyx.mobiledgex.net.csr signed by out/mex-ca.key

The DME can now be run with --tls ./out/dme.xyz.mobiledgex.net.crt

IP address based:
$ certstrap request-cert --ip 127.0.0.1
Enter passphrase (empty for no passphrase): 
<leave blank>
Enter same passphrase again: 
<leave blank>
Created out/127.0.0.1.key
Created out/127.0.0.1.csr

$ certstrap sign --CA mex-ca 127.0.0.1

The DME can now be run with --tls ./out/127.0.0.1.crt


Server certs must be generated for every IP or domain through which clients will access.

Generating Client Certs:

Client certs can be shared for all clients.

$ certstrap request-cert --domain mex-client

Enter passphrase (empty for no passphrase): 
<leave blank>
Enter same passphrase again: 
<leave blank>

Created out/mex-client.key
Created out/mex-client.csr

$ certstrap sign --CA mex-ca mex-client
Created out/mex-client.crt from out/mex-client.csr signed by out/mex-ca.key 


Running DME example:
dme-server   --tls ./out/dme.xyz.mobiledgex.net.crt
2018-08-19T22:09:57.386-0500    INFO    dme-server/dme-notify.go:36     notify client to        {"addrs": "127.0.0.1:50001"}
Loading certfile ./out/dme.xyz.mobiledgex.net.crt cafile out/mex-ca.crt keyfile ./out/dme.xyz.mobiledgex.net.key

Running edgectl client example:
$ edgectl --addr  dme.xyz.mobiledgex.net:50051 dme RegisterClient --tls ./out/mex-client.crt                        
using TLS credentials server dme.xyz.mobiledgex.net certfile ./out/mex-client.crt keyFile ./out/mex-client.key
status: ME_SUCCESS
sessioncookie: eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MzQ4MjExNjUsImp0aSI6IjEyNy4wLjAuMSIsImlhdCI6MTUzNDczNDc2NX0.wNIXQhUvS1IBqOEV6CA0QPdR1iW1-oKqs_Kux3WJIzD4FvhODBwIfur0yWC68UfRBi4_72mqbxQxKKOFmO7T3IEMmEktQ-Uu0VS7GNVX05hRsVVj6iipp_cnBZEOyrlbauTS6dif0XJ_QZLIVbocZoHD-M3T-EO58Gqo46hLP6U2bg_diAZ6i7VM4PLrETeH6W4V0DlBVj9T35p2wi2hUeLj3AtlV3EbVl3v3xTwEXFsnxn7ol4vgWQ7BmB4N4C7HWWMsdCfJ6DgbusgdO3snXBkQtGuZ3qWkqi0TDtAaDWW37WNifOSdOsrCNYWSbbR4jelrP5KitvouX-p51F3auViaAC8VKh_-2G2kN_279APy8Qx-4k0EleB8_7qLJmKGaYqgRbaBqSybM-vbCJNXAhL3yASS-I-vorvPJd6PFryFMT4PdUzeNZLuGXu5FH2td6NpU0QooIrfO6PT4-8DU0U63pr4ebEjttJJI39VzVmebbsj8Jeui79J5Ldzor9WYn9NFPdAn9gvRWUv_WrniaEP4XNDb4nsNl5bsStp8zmI58zBpoquUapiE6ND0W7py2MsC-pwYHk-TM_knisA0S_tejSvVtAINmQ_6nz6Yu_w-L1YIA255IW9kG2HVVHOr7K_EjJgmHAnTo8IcAd3uslgRfJMB904zVIytezRz8

