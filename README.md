[![DOI](https://zenodo.org/badge/656730442.svg)](https://zenodo.org/badge/latestdoi/656730442)


# e2e-slicing-in-a-box

# Introduction
End-to-End network slicing (RAN, Transport, Edge Cloud) using ContainerNet, NS-3, DinD, and UERANSIM.
Based on [lasseufpa / emulation-setup-networking](https://github.com/lasseufpa/emulation-setup-networking).

# Requirements
Tested with **Ubuntu 18.04.**

# Dependencies:
1. ContainerNet
2. ONOS
3. NS-3 
4. UERANSIM
5. python3 & pip
6. Docker
7. Tmux

# Installation

## ONOS and ContainerNet
First, make sure you have python3 with pip installed.

`install.py` will try to install all dependencies needed in a `tools/` folder. Everything should work, but if something goes wrong, you will need to install the dependencies that failed in `tools/` (or in the system).

Run install.py:

```console
python3 install.py
```

# How to use & Configuration

To start the emulation setup, use:

``` console
sudo ./start.sh
```

`start.sh` will start Tmux and run ONOS, Containernet and Free5GC with our topology.

After 1 minute, verify on the `onos-cli` window that all apps cited in `controller/onoscmd` were activated. You should also check the `onos` and `free5gc` windows to see if any container stoped working.

After checking all above, run the following commands on the `scenario` window:

``` console
./create_intents.sh             # to install routing on each switch
./connect_core.sh               # to connect the 5G core containers to Containernet
./start_ran.sh                  # to connect the UE to the respective switch
./ns_command_build.sh           # to start NS-3 and connect it to the gNB and UE
```

## Custom network topology

You can emulate any network topology with containernet, which uses the same interface as Mininet.

For further information on implementing custom network topologies on Mininet, please check Mininet [documentation](https://github.com/mininet/mininet/wiki/Introduction-to-Mininet#creating-topologies).

We also provide an example of a custom network topology, the NSFNet topology. You can check it out at `src/topologies/nsfnet.py`. You can use it with `run_demo.py` to take a look.

We recommend using the following convention when creating a new network topology:

- `hX` for hosts, where X is an integer between 1 and the number of hosts.
- `sX` for switches, where X is an integer between 1 and the number of switches.
- `IP(hX)` = `10.0.0.X`, the IP address of a host `hX` is directly related to the host id `X`.
- `mac(hX)` = `00:00:00:00:00:X`, the mac address of a host `hX` is directly related to the host id `X` in hexadecimal.

PS: For now, our routing module only supports a maximum of 255 hosts.


## free5GC installation

## Prerequisites

- [GTP5G kernel module](https://github.com/free5gc/gtp5g): needed to run the UPF
- [Docker Engine](https://docs.docker.com/engine/install): needed to run the Free5GC containers
- [Docker Compose v2](https://docs.docker.com/compose/install): needed to bootstrap the free5GC stack
- [ns-3](https://github.com/lasseufpa/e2e-slicing-in-a-box/blob/main/ns_install.sh): Required for network simulation

- [free5gc-compose](https://github.com/free5gc/free5gc-compose)

### Changing docker compose file
For the project to work, it was necessary to modify the "docker-compose" file and the configuration folder present in the Free5GC repository by the [docker-compose](https://github.com/lasseufpa/e2e-slicing-in-a-box/blob/ns-free5gc/free5gc/docker-compose.yaml) and [config](https://github.com/lasseufpa/e2e-slicing-in-a-box/tree/ns-free5gc/free5gc/config) generated and made available by this repository  

## Reference
- https://github.com/free5gc/free5gc-compose
- https://github.com/open5gs/nextepc/tree/master/docker
- https://github.com/abousselmi/docker-free5gc
