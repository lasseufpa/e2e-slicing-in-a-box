# free5GC installation

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
