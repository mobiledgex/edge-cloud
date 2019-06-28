#!/bin/bash

curl -k -H "Content-Type: application/json" -X POST -d "@qosreq.txt" https://localhost:38001/v1/getqospositionkpi
