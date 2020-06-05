#!/bin/bash
username=$1
password=$2
echo "user $username"
echo "Create user $username"
salt=$(date +%N)
encryptedPassword=$(perl -e 'print crypt($ARGV[0], salt)' $password)
useradd -m -p "$encryptedPassword" "$username"
echo "$username	ALL=(ALL:ALL) ALL" > /etc/sudoers
cp {bin,etc,home,lib,lib32,lib64,libx32root,sbin,srv,usr,var} tmpfs/ -r
