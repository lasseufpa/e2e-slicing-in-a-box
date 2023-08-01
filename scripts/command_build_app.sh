#!/bin/bash

docker compose up > /dev/null &
sleep 20
# Creating both taps 

ip tuntap add tap-left mode tap
ip tuntap add tap-right mode tap

# Making both packet listeners, even if the mac address is not

ip link set tap-left promisc on
ip link set tap-right promisc on

# Upping both of tap's

ip link set tap-right up 
ip link set tap-left up

# Creating both bridges

ip link add name br-left type bridge
ip link add name br-right type bridge

# Making both Tap's linked to the correct bridge

ip link set tap-left master br-left
ip link set tap-right master br-right 

# Creating 'bridge rules"

echo "Defining Bridge Rules"
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-left -p icmp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-right -p icmp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-left -p tcp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-right -p tcp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-left -p udp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-right -p udp -j ACCEPT
echo "Defined"

pid_left=$(docker inspect --format '{{ .State.Pid }}' ueransim-ue)
pid_right=$(docker inspect --format '{{ .State.Pid }}' ueransim)

#Create a new veth pair's to link the right container to the br-right
gNB="ueransim"
ip link add eth_gNB type veth peer name eth
ip link set eth_gNB master br-right
ip link set eth_gNB up
ip link set eth netns $pid_right
docker exec $gNB ip addr add 10.1.1.1/24 dev eth 
docker exec $gNB ip link set eth up
#docker exec -d $gNB ./nr-gnb -c ./config/gnbcfg.yaml > /dev/null &
#sleep 5

#Create a new veth pair's to link the left container to the br-left
ue="ueransim-ue"
ip link add eth_ue type veth peer name eth
ip link set eth_ue master br-left
ip link set eth_ue up
ip link set eth netns $pid_left
docker exec $ue ip addr add 10.1.1.2/24 dev eth
docker exec $ue ip link set eth up
docker exec $ue ip link set dev eth address "52:9c:58:1e:ad:ec"
#docker exec -d $ue ./nr-ue -c ./config/uecfg.yaml > /dev/null &
#sleep 5

# "Upping' the bridges

ip link set br-left up
ip link set br-right up

