#!/bin/sh
sudo /bin/mkdir /ud
sudo /bin/mount -t vfat /dev/$1 /ud
sudo /home/pi/firmware_update
sync
