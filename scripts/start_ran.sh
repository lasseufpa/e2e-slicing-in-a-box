#!/bin/bash
sudo ovs-docker add-port s1 eth10 ueransim --macaddress=00:00:00:00:01:02 --ipaddress=10.100.200.100/24
sudo docker exec -d ueransim ./nr-gnb -c ./config/gnbcfg.yaml
sudo docker exec -d ueransim-ue ./nr-ue -c ./config/uecfg.yaml