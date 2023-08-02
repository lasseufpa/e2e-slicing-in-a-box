# Switches
export  ID_SW1="of:0000000000000001"
export  ID_SW2="of:0000000000000002"
export  ID_SW3="of:0000000000000003"
export  ID_SW4="of:0000000000000004"
export  ID_SW5="of:0000000000000005"
export  ID_SW6="of:0000000000000006"
export ID_EDGE="of:0000000000000007"
export  ID_REG="of:0000000000000008"
export ID_CORE="of:0000000000000009"
export  ID_REM="of:000000000000000a"

# Connection count
export PORTS_S1=1
export PORTS_S2=1
export PORTS_S3=1
export PORTS_S4=1
export PORTS_S5=1
export PORTS_S6=1
export PORTS_EDGE=1
export PORTS_REG=1
export PORTS_CORE=1
export PORTS_REM=1

# Direct connections
export S1_S3="${ID_SW1}/${PORTS_S1}" ; let PORTS_S1++
export S3_S1="${ID_SW3}/${PORTS_S3}" ; let PORTS_S3++
export S2_S4="${ID_SW2}/${PORTS_S2}" ; let PORTS_S2++
export S4_S2="${ID_SW4}/${PORTS_S4}" ; let PORTS_S4++
export S3_S5="${ID_SW3}/${PORTS_S3}" ; let PORTS_S3++
export S5_S3="${ID_SW5}/${PORTS_S5}" ; let PORTS_S5++
export S4_S6="${ID_SW4}/${PORTS_S4}" ; let PORTS_S4++
export S6_S4="${ID_SW6}/${PORTS_S6}" ; let PORTS_S6++

# Crossed connections
export S1_S4="${ID_SW1}/${PORTS_S1}" ; let PORTS_S1++
export S4_S1="${ID_SW4}/${PORTS_S4}" ; let PORTS_S4++
export S2_S3="${ID_SW2}/${PORTS_S2}" ; let PORTS_S2++
export S3_S2="${ID_SW3}/${PORTS_S3}" ; let PORTS_S3++
export S3_S6="${ID_SW3}/${PORTS_S3}" ; let PORTS_S3++
export S6_S3="${ID_SW6}/${PORTS_S6}" ; let PORTS_S6++
export S4_S5="${ID_SW4}/${PORTS_S4}" ; let PORTS_S4++
export S5_S4="${ID_SW5}/${PORTS_S5}" ; let PORTS_S5++

# Connections to the datacenters
export S1_EDGE="${ID_SW1}/${PORTS_S1}" ; let PORTS_S1++
export EDGE_S1="${ID_EDGE}/${PORTS_EDGE}" ; let PORTS_EDGE++
export S2_EDGE="${ID_SW2}/${PORTS_S3}" ; let PORTS_S2++
export EDGE_S2="${ID_EDGE}/${PORTS_EDGE}" ; let PORTS_EDGE++
export S3_REG="${ID_SW3}/${PORTS_S3}" ; let PORTS_S3++
export REG_S3="${ID_REG}/${PORTS_REG}" ; let PORTS_REG++
export S4_REG="${ID_SW4}/${PORTS_S4}" ; let PORTS_S4++
export REG_S4="${ID_REG}/${PORTS_REG}" ; let PORTS_REG++
export S5_CORE="${ID_SW5}/${PORTS_S5}" ; let PORTS_S5++
export CORE_S5="${ID_CORE}/${PORTS_CORE}" ; let PORTS_CORE++
export S6_CORE="${ID_SW6}/${PORTS_S6}" ; let PORTS_S6++
export CORE_S6="${ID_CORE}/${PORTS_CORE}" ; let PORTS_CORE++

# Connections to remote datacenters
export S5_REM="${ID_SW5}/${PORTS_S5}" ; let PORTS_S5++
export REM_S5="${ID_REM}/${PORTS_REM}" ; let PORTS_REM++
export S6_REM="${ID_SW6}/${PORTS_S6}" ; let PORTS_S6++
export REM_S6="${ID_REM}/${PORTS_REM}" ; let PORTS_REM++

# test hosts
export S1_H1="${ID_SW1}/${PORTS_S1}" ; let PORTS_S1++
export REM_H2="${ID_REM}/${PORTS_REM}" ; let PORTS_REM++
export S4_H3="${ID_SW4}/${PORTS_S4}" ; let PORTS_S4++

# Connection to Free5GC
export CORE_5GC="${ID_CORE}/${PORTS_CORE}" ; let PORTS_CORE++

# Connection to UE1 (SW1)
export S1_UE1="${ID_SW1}/${PORTS_S1}" ; let PORTS_S1++

echo Finished importing variables