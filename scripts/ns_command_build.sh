#cd free5gc-compose/
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
TOOLS_DIR="$(dirname $SCRIPT_DIR)/tools"

cd $SCRIPT_DIR
sudo ./command_build_app.sh

cd $TOOLS_DIR/ns-3-dev
./ns3 run vs-e2e  > /dev/null &

gNB="ueransim"
sudo docker exec $gNB ./nr-gnb -c ./config/gnbcfg.yaml &
sudo docker exec $gNB ip route del default
sudo docker exec $gNB ip route add 10.1.1.0/24 via 10.1.2.1 dev eth
sudo docker exec $gNB ip route add default via 10.1.2.2 dev eth
sudo docker exec $gNB ip route del 10.100.200.0/24 dev eth0
sudo docker exec $gNB ip route #To show whether the routes were correctly set

sleep 5
ue="ueransim-ue"
sudo docker exec $ue ip route add 10.1.2.0/24 via 10.1.1.1 dev eth
sudo docker exec -d $ue ./nr-ue -c ./config/uecfg.yaml &


