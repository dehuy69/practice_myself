arm-himix100-linux-gcc fifoserver_origin.c -o fifoserver
arm-himix100-linux-gcc fifoclient_origin.c -o fifoclient
cp fifoserver /mnt/testnfs/ev200_dev/
cp fifoclient /mnt/testnfs/ev200_dev/