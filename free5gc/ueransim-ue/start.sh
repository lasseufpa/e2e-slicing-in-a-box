#!/bin/bash

# Iniciar o comando ./nr-gnb em segundo plano
./nr-gnb -c ./config/gnbcfg.yaml &

# Iniciar o comando ./nr-ue em segundo plano
./nr-ue -c ./config/uecfg.yaml &

# Manter o script em execução para que o contêiner não seja encerrado
tail -f /dev/null