#!/bin/bash

OUTDIR=/tmp/test_results
SETUPFILE=setups/local_multi.yml

if [ $# -eq 1 ]; then
   SETUPFILE=$1
fi

echo "using setupfile $SETUPFILE"

for test in `cat tests.txt`;
do
   fname=`echo $test |cut -d "." -f 1`
   out=$OUTDIR/$fname
   echo "Running test $test against setup $SETUPFILE, output in $out"
   cmd="e2e-tests -testfile testfiles/$test -outputdir $out -setupfile $SETUPFILE -datadir data "
   echo $cmd
   $cmd|grep -e Summary -e PASS -e FAIL 
 
done
