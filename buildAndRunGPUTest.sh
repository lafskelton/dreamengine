#!/bin/bash
cd pkg/neural
#
printf "\n\nCUDA BUILD\n\n"
rm lib/libmodeldef.so 
nvcc --ptxas-options=-v --compiler-options '-fPIC' -g -G --library-path='/usr/local/cuda-11.2/lib64'  -o lib/libmodeldef.so --shared modeldef.cu
printf "\n"
#
printf "GO BUILD\n\n"
go build ../../test/gpuTest/main.go
printf "\n"
rm ../../bin/gpuTest
mv main ../../bin/gpuTest
#
sudo cp /usr/local/cuda-11.2/lib64/libmodeldef.so ../../libmodeldef_OLD.so
sudo rm /usr/local/cuda-11.2/lib64/libmodeldef.so
sudo cp lib/libmodeldef.so /usr/local/cuda-11.2/lib64/libmodeldef.so
#
printf "\nRUN TEST\n\n"
./../../bin/gpuTest
#
printf "\n"