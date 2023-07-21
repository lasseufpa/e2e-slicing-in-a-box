#!/bin/bash
sudo ovs-docker add-port s1 eth0 uegnb --ipaddress=10.100.200.100/24
sudo docker exec -d uegnb ./nr-gnb -c ./config/gnbcfg.yaml
sudo docker exec -d uegnb ./nr-ue -c ./config/uecfg.yaml