sudo ovs-docker add-port s1 eth0 ueransim --ipaddress=10.100.200.100/24
sudo ovs-docker add-port s1 eth0 ueransim-ue --ipaddress=10.100.200.101/24
sudo ovs-vsctl add-port s2 br-free5gc
