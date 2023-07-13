cd free5gc-compose/
sudo ./command_build_app.sh
cd ../
./ns3 run src/tap-bridge/examples/tap-wifi-virtual-machine.cc > /dev/null &
gNB="ueransim"
docker exec -d $gNB ./nr-gnb -c ./config/gnbcfg.yaml > /dev/null &
docker exec -d $gNB ip route del default via 10.100.200.1 dev eth0
docker exec -d $gNB ip route add default via 10.1.1.1 dev eth

sleep 5
ue="ueransim-ue"
docker exec -d $ue ./nr-ue -c ./config/uecfg.yaml > /dev/null &

sleep 330
cd free5gc-compose/
docker compose stop