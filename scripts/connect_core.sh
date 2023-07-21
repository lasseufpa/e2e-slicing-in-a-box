sudo ovs-vsctl add-port s2 br-free5gc
sudo docker exec -d upf ip route add 10.0.0.0/24 dev eth0