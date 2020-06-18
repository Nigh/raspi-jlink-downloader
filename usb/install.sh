#!/bin/sh
cp -f 100-add-usb.rules /etc/udev/rules.d/100-add-usb.rules
cp -f mount-usb.sh /sbin/mount-usb.sh
cp -f umount-usb.sh /sbin/umount-usb.sh
cp -f firmware_update /home/pi/firmware_update
cp -f jlink_downloader /home/pi/jlink_downloader
cp -f firmware.enc /boot
chmod +x /sbin/mount-usb.sh
chmod +x /sbin/umount-usb.sh
chmod +x /home/pi/firmware_update
chmod +x /home/pi/jlink_downloader
