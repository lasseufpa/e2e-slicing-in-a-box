#cd free5gc-compose/
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
TOOLS_DIR="$(dirname $SCRIPT_DIR)/tools"

cd $SCRIPT_DIR
sudo ./command_build_app.sh

cd $TOOLS_DIR/ns-allinone-3.38/ns-3.38
./ns3 run src/tap-bridge/examples/tap-wifi-virtual-machine.cc > /dev/null &

export gNB="ueransim"
docker exec -d $gNB ./nr-gnb -c ./config/gnbcfg.yaml > /dev/null &
docker exec -d $gNB ip route del default
docker exec -d $gNB ip route add default via 10.1.1.1 dev eth
docker exec -d $gNB ip route del 10.100.200.0/24 dev eth0
docker exec -d $gNB ip route #To show whether the routes were correctly set

sleep 5
ue="ueransim-ue"
docker exec -d $ue ./nr-ue -c ./config/uecfg.yaml > /dev/null &

#sleep 330
#cd free5gc-compose/
#docker compose stop