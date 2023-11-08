#!/bin/bash

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
COMPOSE_DIR="$(dirname $SCRIPT_DIR)/docker-compose"



# Creating both taps 
ip tuntap add tap-ue mode tap
ip tuntap add tap-gnb mode tap

# Making both packet listeners, even if the mac dest address is not from the tap 
ip link set tap-ue promisc on
ip link set tap-gnb promisc on

# Upping both of tap's
ip link set tap-gnb up 
ip link set tap-ue up

# Creating both bridges
ip link add name br-ue type bridge
ip link add name br-gnb type bridge

# Making both Tap's linked to the correct bridge
ip link set tap-ue master br-ue
ip link set tap-gnb master br-gnb 

# Creating 'bridge rules"
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-ue -p icmp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-gnb -p icmp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-ue -p tcp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-gnb -p tcp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-ue -p udp -j ACCEPT
sudo iptables -I FORWARD -m physdev --physdev-is-bridged -i br-gnb -p udp -j ACCEPT

pid_ue=$(docker inspect --format '{{ .State.Pid }}' ueransim-ue)
pid_gnb=$(docker inspect --format '{{ .State.Pid }}' ueransim)

#Create a new veth pair's to link the gnb container to the br-gnb
gNB="ueransim"
ip link add eth-gnb type veth peer name eth
ip link set eth-gnb master br-gnb
ip link set eth-gnb up
ip link set eth netns $pid_gnb
docker exec $gNB ip addr add 10.1.2.2/24 dev eth 
docker exec $gNB ip link set eth up
#docker exec -d $gNB ./nr-gnb -c ./config/gnbcfg.yaml > /dev/null &
#sleep 5

#Create a new veth pair's to link the ue container to the br-ue
ue="ueransim-ue"
ip link add eth-ue type veth peer name eth
ip link set eth-ue master br-ue
ip link set eth-ue up
ip link set eth netns $pid_ue
docker exec $ue ip addr add 10.1.1.2/24 dev eth
docker exec $ue ip link set eth up
docker exec $ue ip link set dev eth address "52:9c:58:1e:ad:ec"
#docker exec -d $ue ./nr-ue -c ./config/uecfg.yaml > /dev/null &
#sleep 5

# "Upping' the bridges
ip link set br-ue up
ip link set br-gnb up

# Routing Table
sudo docker exec -d $gNB ip route del default
sudo docker exec -d $gNB ip route add 10.1.1.0/24 via 10.1.2.1 dev eth
sudo docker exec -d $gNB ip route del 10.100.200.0/24 dev eth0
sudo docker exec -d $gNB ip route del 10.100.200.0/24 dev eth10 proto kernel scope link src 10.100.200.100 
sudo docker exec -d $gNB ip route add 10.100.200.0/24 dev eth10 
sudo docker exec -d $ue ip route add 10.100.200.100/24 via 10.1.1.1 dev eth
sudo docker exec -d $ue ip route add 10.1.2.0/24 via 10.1.1.1 dev eth

