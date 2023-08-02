ONOS_SSH="sshpass -p karaf ssh -n -p 8101 -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no karaf@localhost"
ONOS_CMD_FILE="onossdn"
ONOS_ENV="onosdevices"

source $ONOS_ENV

# READ onos commands
while IFS= read -r p 
do
    if [[ ! -z $p ]] && [[ $p != \#* ]];
    then
        #echo "$p"
        com=$(eval echo "$p")
        echo "$p" "--->>>" "$com"
        ${ONOS_SSH} $com
        sleep 1;
    fi
done < "${ONOS_CMD_FILE}"