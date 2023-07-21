#!/bin/bash

source $HOME/.bashrc

if [ ! -f ~/.tmux.conf ];
then
  cp .tmux.conf ~/
fi

# set paths
SFLOW_DIR="$PWD/tools/sflow-rt"
SETUP_DIR="$PWD"

source ${SETUP_DIR}/.env

dir_exists () {
  if [ ! -d $1 ]; then
    echo "Directory $1 DOES NOT exists."
    exit
  fi
}

#dir_exists $SFLOW_DIR
dir_exists $SETUP_DIR

# change this to your liking
PROJECT_NAME=emulation

# do not change this
MAIN_DIR=~/"tmux_logs"

if [ ! -d $MAIN_DIR ]; then
  mkdir -p $MAIN_DIR/$PROJECT_NAME
fi

# following commands will be executed first, in each window
# pre_input=""

# define commands
# 'name' 'command'
# DO NOT PUT spaces in the names
input=(
  'onos' "sudo docker compose -f ./docker-compose/onos.yaml up
  "
  'onos-cli'  "echo 'Waiting for ONOS to start' ; sleep 30 ;
               sshpass -p karaf ssh -p 8101 -o StrictHostKeyChecking=no karaf@localhost app activate ${ONOS_DEFAULT_APPS} ${ONOS_APPS_CUSTOM} ; sshpass -p karaf ssh -p 8101 -o StrictHostKeyChecking=no karaf@localhost
  "
  'free5gc' "sudo docker compose -f ./tools/free5gc-compose/free5gc.yaml up
  "
  'ueransim' "sudo docker compose -f ./config/uegnb.yaml up
  "
  'scenario' ""
)

init_window="scenario"

###########################
### DO NOT MODIFY BELOW ###
###########################

SESSION_NAME=testbed

# prefere the user-compiled tmux
if [ -f /usr/local/bin/tmux ]; then
  export TMUX_BIN=/usr/local/bin/tmux
else
  export TMUX_BIN=/usr/bin/tmux
fi

# find the session
FOUND=$( $TMUX_BIN ls | grep $SESSION_NAME )

if [ $? == "0" ]; then

  echo "The session already exists"
  exit
fi

# Absolute path to this script. /home/user/bin/foo.sh
SCRIPT=$(readlink -f $0)
# Absolute path this script is in. /home/user/bin
SCRIPTPATH=`dirname $SCRIPT`

if [ -z ${TMUX} ];
then
  TMUX= $TMUX_BIN new-session -s "$SESSION_NAME" -d
  echo "Starting new session."
else
  echo "Already in tmux, leave it first."
  exit
fi

# get the iterator
ITERATOR_FILE="$MAIN_DIR/$PROJECT_NAME"/iterator.txt
if [ -e "$ITERATOR_FILE" ]
then
  ITERATOR=`cat "$ITERATOR_FILE"`
  ITERATOR=$(($ITERATOR+1))
else
  echo "iterator.txt does not exist, creating it"
  touch "$ITERATOR_FILE"
  ITERATOR="0"
fi
echo "$ITERATOR" > "$ITERATOR_FILE"

# create file for logging terminals' output
LOG_DIR="$MAIN_DIR/$PROJECT_NAME/"
SUFFIX=$(date +"%Y_%m_%d_%H_%M_%S")
SUBLOG_DIR="$LOG_DIR/"$ITERATOR"_"$SUFFIX""
TMUX_DIR="$SUBLOG_DIR/tmux"
mkdir -p "$SUBLOG_DIR"
mkdir -p "$TMUX_DIR"

# link the "latest" folder to the recently created one
rm "$LOG_DIR/latest"
rm "$MAIN_DIR/latest"
ln -sf "$SUBLOG_DIR" "$LOG_DIR/latest"
ln -sf "$SUBLOG_DIR" "$MAIN_DIR/latest"

# create arrays of names and commands
for ((i=0; i < ${#input[*]}; i++));
do
  ((i%2==0)) && names[$i/2]="${input[$i]}"
  ((i%2==1)) && cmds[$i/2]="${input[$i]}"
done

# run tmux windows
for ((i=0; i < ${#names[*]}; i++));
do
  $TMUX_BIN new-window -t $SESSION_NAME:$(($i+1)) -n "${names[$i]}"
done

sleep 3

# start loggers
for ((i=0; i < ${#names[*]}; i++));
do
  $TMUX_BIN pipe-pane -t $SESSION_NAME:$(($i+1)) -o "ts | cat >> $TMUX_DIR/$(($i+1))_${names[$i]}.log"
done

# send commands
for ((i=0; i < ${#cmds[*]}; i++));
do
  $TMUX_BIN send-keys -t $SESSION_NAME:$(($i+1)) "cd $SCRIPTPATH;
${cmds[$i]}"
done

# identify the index of the init window
init_index=0
for ((i=0; i < ((${#names[*]})); i++));
do
  if [ ${names[$i]} == "$init_window" ]; then
    init_index=$(expr $i + 1)
  fi
done

$TMUX_BIN select-window -t ${SESSION_NAME}:$init_index

$TMUX_BIN -2 attach-session -t ${SESSION_NAME}

# clear

