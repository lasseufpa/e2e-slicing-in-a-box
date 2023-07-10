from mininet.net import Mininet
from mininet.link import TCLink
from mininet.node import RemoteController
from mininet.topo import SingleSwitchTopo, Topo
from mininet.cli import CLI

from src.routing.mininet_based.routing import StaticRouter 
from src.topologies.nsfnet import NSFNet

# Compile and run sFlow helper script
# - configures sFlow on OVS
# - posts topology to sFlow-RT

with open("tools/sflow-rt/extras/sflow.py") as f:
    exec(f.read())

def main():
    ### Init network

    #topo = SingleSwitchTopo(2)
    #topo = NSFNet() # Our NSFNet topology implementation

    topo=Topo()

    #Create Nodes
    topo.addHost("h1")
    topo.addHost("h2")
    topo.addHost("h3")
    topo.addHost("h4")
    topo.addSwitch('s1')
    topo.addSwitch('s2')
    topo.addSwitch('s3')

    #Create links
    topo.addLink('s1','s2')
    topo.addLink('s1','s3')
    topo.addLink('s2','s3')
    topo.addLink('h1','s1')
    topo.addLink('h2','s2')
    topo.addLink('h3','s3')
    topo.addLink('h4','s3')
    
    net = Mininet(topo=topo, link=TCLink, controller=RemoteController)
    net.start()

    ### Routing
    router = StaticRouter(topo)
    router.view() # view network topology
    router.reset() # reset any routing trash from previous runs

    router.route(src=None, dst=None) # every host can reach every other host
    # router.route(dst='h2') # every host can reach only h2
    
    ### Start mininet CLI

    net.pingAll()

    CLI(net)

    ### End network 

    net.stop() 
    router.reset() # reset routing

if __name__ == '__main__':
    main()
