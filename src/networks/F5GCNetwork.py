from mininet.net import Mininet
from mininet.link import TCLink
from mininet.node import RemoteController
from mininet.cli import CLI

class F5GCNetwork():
    def __init__(self, topo):
        self.topo = topo
        self.net = None

    def start(self):
        if self.net == None:
            self.net = Mininet(topo=self.topo, controller=None, link=TCLink, autoSetMacs=True, autoStaticArp=True)
            self.net.addController('c0',controller=RemoteController, ip='127.0.0.1', port=6653)
            self.net.start()

    def stop(self):
        if self.net != None:
            self.net.stop()
            self.net = None

    def CLI(self):
        CLI(self.net)

    def pingAll(self):
        self.net.pingAll()


