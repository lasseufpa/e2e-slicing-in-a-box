from mininet.topo import Topo

# Topology implemented
#  --- s1 ---3m--- s3 ---11ms--- s5 ---70ms
#      |  \        /| \         / |      |
#     .5ms \      /2ms \       / 11ms    |
#      |    3ms  /  |   11ms  /   |      |
#     edge     \/  reg      \/   core  remote ---
#      |     3ms\   |    11ms\    |      |
#     .5ms  /    \ 2ms  /     \  11ms    |
#      |   /      \ |  /       \  |      |
#      s2 ---3m--- s4 ---11ms--- s6 ---70ms
#                   |
#                   |


class F5GCTopo(Topo):
    def build(self):
        ### Init network

        ### Switches
        s1  = self.addSwitch('s1') # access
        s2  = self.addSwitch('s2') # access
        s3  = self.addSwitch('s3')
        s4  = self.addSwitch('s4')
        s5  = self.addSwitch('s5')
        s6  = self.addSwitch('s6')
        edge  = self.addSwitch('s7')
        regional  = self.addSwitch('s8')
        core  = self.addSwitch('s9')
        remote = self.addSwitch('s10')
        
        # Direct connections
        self.addLink(s1, s3, delay='3ms', bw=1000)
        self.addLink(s2, s4, delay='3ms', bw=1000)
        self.addLink(s3, s5, delay='11ms', bw=1000)
        self.addLink(s4, s6, delay='11ms', bw=1000)
        # Crossed connections
        self.addLink(s1, s4, delay='3ms', bw=1000)
        self.addLink(s2, s3, delay='3ms', bw=1000)
        self.addLink(s3, s6, delay='11ms', bw=1000)
        self.addLink(s4, s5, delay='11ms', bw=1000)

        # Connections to the datacenters
        self.addLink(s1, edge, delay='0.5ms', bw=1000)
        self.addLink(s2, edge, delay='0.5ms', bw=1000)
        self.addLink(s3, regional, delay='2ms', bw=1000)
        self.addLink(s4, regional, delay='2ms', bw=1000)
        self.addLink(s5, core, delay='11ms', bw=1000)
        self.addLink(s6, core, delay='11ms', bw=1000)

        # Connections to remote datacenters
        self.addLink(s5, remote, delay='70ms', bw=1000)
        self.addLink(s6, remote, delay='70ms', bw=1000)

        # test hosts
        h1 = self.addHost('h1')
        h2 = self.addHost('h2')
        h3 = self.addHost('h3')
        self.addLink(h1,s1)
        self.addLink(h2,remote)
        self.addLink(h3,s4)


