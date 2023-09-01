#!/bin/bash
sudo ovs-docker add-port s1 eth10 ueransim --ipaddress=10.100.200.100/24 --macaddress=00:00:00:00:01:02 
echo "gNB connected!"