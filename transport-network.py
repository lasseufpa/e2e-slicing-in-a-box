#!/usr/bin/python
"""
execute docker pull ubuntu:trusty
"""
from mininet.net import Containernet
from mininet.node import Controller
from mininet.cli import CLI
from mininet.link import TCLink
from mininet.log import info, setLogLevel
setLogLevel('info')

net = Containernet(controller=Controller)
info('*** Adding controller\n')
net.addController('c0')

info('*** Adding docker containers\n')

gNodeB1 = net.addDocker('gnb1', ip='10.0.0.2', dimage="ubuntu:trusty")
gNodeB2 = net.addDocker('gnb2', ip='10.0.0.3', dimage="ubuntu:trusty")

bbudc = net.addDocker('bbudc',ip='10.0.0.4', dimage="ubuntu:trusty")
regionaldc = net.addDocker('regionaldc', ip='10.0.0.5', dimage="ubuntu:trusty")
coredc = net.addDocker('coredc',ip='10.0.0.6', dimage="ubuntu:trusty")

info('*** Adding switches\n')
s1 = net.addSwitch('s1')
s2 = net.addSwitch('s2')
s3 = net.addSwitch('s3')
s4 = net.addSwitch('s4')
s5 = net.addSwitch('s5')
s6 = net.addSwitch('s6')
info('*** Creating links\n')
net.addLink(gNodeB1, s1)
net.addLink(gNodeB2, s2)
net.addLink(s1, bbudc, cls=TCLink, delay='0.5ms', bw=1000)
net.addLink(s2, bbudc, cls=TCLink, delay='0.5ms', bw=1000)
net.addLink(s1, s3, cls=TCLink, delay='3ms', bw=1000)
net.addLink(s2, s4, cls=TCLink, delay='3ms', bw=1000)
net.addLink(s3, regionaldc, cls=TCLink, delay='2ms', bw=1000)
net.addLink(s4, regionaldc, cls=TCLink, delay='2ms', bw=1000)
net.addLink(s3, s5, cls=TCLink, delay='11ms', bw=1000)
net.addLink(s4, s6, cls=TCLink, delay='11ms', bw=1000)
net.addLink(s5, coredc, cls=TCLink, delay='11ms', bw=1000)
net.addLink(s6, coredc, cls=TCLink, delay='11ms', bw=1000)

net.addLink(s1, s4, cls=TCLink, delay='3ms', bw=1000)
net.addLink(s2, s3, cls=TCLink, delay='3ms', bw=1000)
net.addLink(s3, s6, cls=TCLink, delay='11ms', bw=1000)
net.addLink(s4, s5, cls=TCLink, delay='11ms', bw=1000)

info('*** Starting network\n')
net.start()
info('*** Testing connectivity\n')
net.ping([gNodeB1, gNodeB2])
info('*** Running CLI\n')
CLI(net)
info('*** Stopping network')
net.stop()

