ONOS_SSH="sshpass -p karaf ssh -n -p 8101 -o StrictHostKeyChecking=no karaf@localhost"
ONOS_CMD_FILE="onossdn"
ONOS_ENV="onosdevices"

source $ONOS_ENV

#IFS="" read -r -a lines < $ONOS_CMD_FILE || (( ${#onoscommands[@]} ))
#echo ${#onoscommands[@]}

while IFS= read -r p 
#for p in "${#onoscommands[@]}"
do
    #echo "$p"
    com=$(eval echo "$p")
    echo "$p" "--->>>" "$com"
    ${ONOS_SSH} $com
    #sleep 1;
    #echo next
done < "${ONOS_CMD_FILE}"