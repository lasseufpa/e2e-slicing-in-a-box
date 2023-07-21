import atexit
from logging import info
import os
import subprocess
from mininet.net import Mininet
from mininet.link import TCLink
from mininet.node import Docker, RemoteController
from mininet.topo import Topo
from mininet.cli import CLI
from mininet.log import info,setLogLevel

def run_command(command):
    returncode = subprocess.run(command.split()).returncode
    if returncode != 0:
        print(f'failed to run:\n{command}')
        exit(0)

setup_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), 'scripts')

net = None
#router = None

def main():
    ### Init network

    topo=Topo()

    info('*** Adding docker containers\n')

    gNodeB1 = topo.addHost('h1', cls=Docker, ip='10.0.0.2', dimage="ubuntu:trusty")
    gNodeB2 = topo.addHost('h2', cls=Docker, ip='10.0.0.3', dimage="ubuntu:trusty")

    bbudc = topo.addHost('h3',cls=Docker, ip='10.0.0.4', dimage="ubuntu:trusty")
    regionaldc = topo.addHost('h4', cls=Docker, ip='10.0.0.5', dimage="ubuntu:trusty")
    coredc = topo.addHost('h5',cls=Docker, ip='10.0.0.6', dimage="ubuntu:trusty")

    info('*** Adding switches\n')
    s1 = topo.addSwitch('s1')
    s2 = topo.addSwitch('s2')
    s3 = topo.addSwitch('s3')
    s4 = topo.addSwitch('s4')
    s5 = topo.addSwitch('s5')
    s6 = topo.addSwitch('s6')
    info('*** Creating links\n')
    topo.addLink(gNodeB1, s1)
    topo.addLink(gNodeB2, s2)
    topo.addLink(s1, bbudc, delay='0.5ms', bw=1000)
    topo.addLink(s2, bbudc, delay='0.5ms', bw=1000)
    topo.addLink(s1, s3, delay='3ms', bw=1000)
    topo.addLink(s2, s4, delay='3ms', bw=1000)
    topo.addLink(s3, regionaldc, delay='2ms', bw=1000)
    topo.addLink(s4, regionaldc, delay='2ms', bw=1000)
    topo.addLink(s3, s5, delay='11ms', bw=1000)
    topo.addLink(s4, s6, delay='11ms', bw=1000)
    topo.addLink(s5, coredc, delay='11ms', bw=1000)
    topo.addLink(s6, coredc, delay='11ms', bw=1000)

    topo.addLink(s1, s4, delay='3ms', bw=1000)
    topo.addLink(s2, s3, delay='3ms', bw=1000)
    topo.addLink(s3, s6, delay='11ms', bw=1000)
    topo.addLink(s4, s5, delay='11ms', bw=1000)

    net = Mininet(topo=topo, controller=None, link=TCLink, autoSetMacs=True, autoStaticArp=True)
    net.addController('c0',controller=RemoteController, ip='127.0.0.1', port=6653)
    net.start()

    info('*** Starting RAN\n')
    os.chdir(setup_path)
    run_command('./start_ran.sh')
    info('*** Connecting Core to the network\n')
    run_command('./connect_core.sh')

    ### Start mininet CLI
    CLI(net)

    ### End network 
    net.stop() 

def stopNetwork():
    if net is not None:
        net.stop()
        #router.reset() # reset routing

if __name__ == '__main__':
    atexit.register(stopNetwork)
    setLogLevel('info')
    main()
