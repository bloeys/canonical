#!/bin/bash

set -e

#Install dependencies
sudo apt-get update
sudo apt-get install \
    make \
    qemu \
    qemu-system-x86 \
    bison \
    build-essential \
    flex \
    libelf-dev \
    libssl-dev \
    ncurses-dev \
    grub-common \
    xorriso \
    grub-pc-bin

#Download linux kernel compressed source and extract
kernelVersion="linux-5.18.10"
kernelTarFile="kernel.tar.xz"

wget -O "$kernelTarFile" "https://cdn.kernel.org/pub/linux/kernel/v5.x/$kernelVersion.tar.xz"
tar xvf "$kernelTarFile"

#Configure and build kernel
make -C "./$kernelVersion" O="./build/x86/" x86_64_defconfig
make -C "./$kernelVersion" O="./build/x86/" -j $(nproc)

#Download busybox
busyboxVersion="busybox-1.35.0"
wget -O "$busyboxVersion.tar.bz2" "https://busybox.net/downloads/$busyboxVersion.tar.bz2"
tar xvf "$busyboxVersion.tar.bz2"

#Configure and statically build busybox
mkdir -p "./$busyboxVersion/build"
make -C "./$busyboxVersion" O="./build/" defconfig

sed -i 's/# CONFIG_STATIC is not set/CONFIG_STATIC=y/g' "./$busyboxVersion/build/.config"
make -C "./$busyboxVersion" O="./build/" -j $(nproc)

#Produce an '_install' directory that contains links from standard unix tools to the main busybox executable
make -C "./$busyboxVersion" O="./build/" install

#Setup initramfs and init script
mkdir -p ./initramfs/busybox/{bin,sbin,etc,proc,sys,usr/{bin,sbin},dev}
cp -a "./$busyboxVersion/build/_install/"* "./initramfs/busybox" #Copy busybox unix tools preserving symlinks

initFile="./initramfs/busybox/init"
cat << EOF > "$initFile"
#!/bin/sh

mount -t proc none /proc
mount -t sysfs none /sys

clear
echo -e "\n\n--------------\nHello Canonical! :D\n"
echo "You are now in a shell. Press ENTER and type away"
echo -e "--------------\n"

exec sh
EOF
chmod +x "$initFile"

#Produce gzip containing initramfs that the kernel expects
#NOTE: This must be built using 'find .', not using a path otherwise kernel will fail to find the files
cd "./initramfs/busybox/"
find . -print0 \
    | cpio --null -ov --format=newc \
    | gzip -9 > "../../initramfs-busybox.cpio.gz"

cd ../../

#Now we have everything we need to boot, so we will package everything into a bootable ISO with grub
mkdir -p "./iso/boot/grub"
cat << EOF > "./iso/boot/grub/grub.cfg"
set timeout=0
set default=0

menuentry "simple-linux" {
    linux /boot/bzImage
    initrd /boot/initramfs-busybox.cpio.gz
}
EOF

cp "./$kernelVersion/build/x86/arch/x86_64/boot/bzImage" "./iso/boot/"
cp "./initramfs-busybox.cpio.gz" "./iso/boot/"
grub-mkrescue -o "simple-linux.iso" "./iso"

#Run!
qemu-system-x86_64 -m "1024M" -cdrom "./simple-linux.iso"
