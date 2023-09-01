#cd free5gc-compose/
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
TOOLS_DIR="$(dirname $SCRIPT_DIR)/tools"



gNB="ueransim"
#sudo docker exec -d $gNB ./nr-gnb -c ./config/gnbcfg.yaml &
sudo docker exec -d $gNB ip route del default
sudo docker exec -d $gNB ip route add 10.1.1.0/24 via 10.1.2.1 dev eth
sudo docker exec -d $gNB ip route del 10.100.200.0/24 dev eth0
sudo docker exec -d $gNB ip route del 10.100.200.0/24 dev eth10 proto kernel scope link src 10.100.200.100 
sudo docker exec -d $gNB ip route add 10.100.200.0/24 dev eth10 
ue="ueransim-ue"
sudo docker exec -d $ue ip route add 10.100.200.100/24 via 10.1.1.1 dev eth
sudo docker exec -d $ue ip route add 10.1.2.0/24 via 10.1.1.1 dev eth
#sudo docker exec -d $ue ./nr-ue -c ./config/uecfg.yaml 


