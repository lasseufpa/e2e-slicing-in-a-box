#!/bin/bash
echo "--------- Connecting ---------"
echo "Connecting gNB"
sudo ovs-docker add-port s1 eth10 ueransim --ipaddress=10.100.200.100/24 --macaddress=00:00:00:00:01:02 
#Verify if the gNB can ping h1 (simple ghost host on Mininet topology)
sudo docker exec -t ueransim ping -c 5 -I eth10 10.100.200.101

if [[ $? -eq 0 ]]; then
    echo "gNB connected"
else
    echo "gNB Not connected, please verify"
    exit 1
fi


echo "Connecting Free5gc Core" 
sudo ovs-vsctl add-port s9 br-free5gc
sudo docker exec -d upf ip route add 10.0.0.0/24 dev eth0
#Verify if the gNB can ping the core (AMF)
sudo docker exec -t ueransim ping -c 5 -I eth10 10.100.200.9
if [[ $? -eq 0 ]]; then
    echo "Core connected"
else
    echo "Core Not connected, please verify"
    exit 1
fi

echo "Connecting Server"
sudo ovs-docker add-port s10 eth10 server --ipaddress=10.100.200.120/24 --macaddress=00:00:00:02:03:04
#Verify if the gNB can ping the Server
sudo docker exec -t ueransim ping -c 5 -I eth10 10.100.200.120
if [[ $? -eq 0 ]]; then
    echo "Server connected"
else
    echo "Server Not connected, please verify"
    exit 1
fi