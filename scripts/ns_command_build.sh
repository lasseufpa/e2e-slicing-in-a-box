#cd free5gc-compose/
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
TOOLS_DIR="$(dirname $SCRIPT_DIR)/tools"

cd $SCRIPT_DIR
sudo ./command_build_app.sh

cd $TOOLS_DIR/ns-allinone-3.38/ns-3.38
./ns3 run src/tap-bridge/examples/tap-wifi-virtual-machine.cc > /dev/null &

gNB="ueransim"
sudo docker exec -d $gNB ./nr-gnb -c ./config/gnbcfg.yaml > /dev/null &
sudo docker exec -d $gNB ip route del default via 10.100.200.1 dev eth0
sudo docker exec -d $gNB ip route add default via 10.1.1.1 dev eth

sleep 5
ue="ueransim-ue"
sudo docker exec -d $ue ./nr-ue -c ./config/uecfg.yaml > /dev/null &

#sleep 330
#cd free5gc-compose/
#docker compose stop