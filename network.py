import atexit
from mininet.log import info,setLogLevel

from src.topologies.F5GCTopo import F5GCTopo
from src.networks.F5GCNetwork import F5GCNetwork

#with open("tools/sflow-rt/extras/sflow.py") as f:
#    exec(f.read())

net = None

def main():
    topo = F5GCTopo()
    net = F5GCNetwork(topo)
    net.start()
    #net.pingAll()
    net.CLI()

def stop():
    if net != None: net.stop()

if __name__ == '__main__':
    atexit.register(stop)
    setLogLevel('info')
    main()