#!/bin/bash
sudo ovs-vsctl add-port s9 br-free5gc
echo "Free5gc connected to swittch s9"
sudo docker exec -d upf ip route add 10.0.0.0/24 dev eth0
