#use with python3
import shlex, subprocess
import shutil
import os
import sys

"""
A simple program to install the dependencies required to run the Testbed
"""

def run_command(command):
    returncode = subprocess.run(command.split()).returncode
    if returncode != 0:
        print(f'failed to run:\n{command}')
        exit(0)

setup_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), 'tools')

if os.path.exists(setup_path) == False:
    print('Creating tools directory!')
    os.mkdir(setup_path)

os.chdir(setup_path)

print("Updating apt")
run_command('sudo apt update')

print("Step 1. Install prerequisites")

run_command('sudo  apt-get -y install ansible  git  aptitude  gcc  python3-dev  libffi-dev  libssl-dev  libxml2-dev  libxslt1-dev  zlib1g-dev  openjdk-8-jre  adduser  libfontconfig1  debian-keyring  debian-archive-keyring  apt-transport-https')
run_command('sudo  apt-get -y install  tmux libjpeg-dev zlib1g-dev python3-pip')
run_command('sudo python3  -m  pip  install  pyyaml')
run_command('sudo python3  -m  pip  install  pandas')
run_command('sudo python3  -m  pip  install  matplotlib')
run_command('sudo python3  -m  pip  install  networkx')
run_command('sudo python3  -m  pip  install  Flask')

run_command('sudo  apt-get -y install  sshpass') # to be able to access onos with password on command line

print("Step 2. Install ContainerNET")

if os.path.exists("containernet") == False:
    run_command('git  clone  https://github.com/containernet/containernet.git')
    os.chdir('containernet/ansible')
    run_command('sudo  ansible-playbook  -i  \"localhost\"  -c  local  install.yml')
    os.chdir(os.path.join(setup_path, 'containernet'))
    run_command('sudo make develop')
    run_command('sudo  apt-get -y install ovn-docker')

run_command('sudo docker pull onosproject/onos')

os.chdir(setup_path)
  
print("Step 3. Install Free5gc")

print("Installing gtp5g kernel module...")
if os.path.exists("gtp5g") == False:
    #installing gtp5g
    run_command('git clone https://github.com/free5gc/gtp5g.git')
    os.chdir("gtp5g")
    run_command('git checkout v0.8.2')
    run_command('make clean')
    run_command('make')
    run_command('sudo make install')

os.chdir(setup_path)

print("Cloning and Updating free5Gc...")
if os.path.exists("free5gc-compose") == False:
    #installing free5gc
    run_command('git  clone  https://github.com/free5gc/free5gc-compose.git')
    run_command('cp ../docker-compose/free5gc.yaml ./free5gc-compose/')
os.chdir(setup_path)



print("Step 4. Install NS3")
print("Installing NS-3 dependencies")
print("Building NS-38 and installing:")
run_command("sudo apt install ccache gir1.2-goocanvas-2.0 python3-gi clang python3-gi-cairo python3-pygraphviz gir1.2-gtk-3.0 ipython3 python3-setuptools\
 qtbase5-dev qtchooser qt5-qmake qtbase5-dev-tools qt5-default openmpi-bin openmpi-common openmpi-doc libopenmpi-dev mercurial gdb valgrind\
 clang-format gsl-bin libgsl-dev libgslcblas0 tcpdump sqlite sqlite3 libsqlite3-dev cmake libc6-dev libc6-dev-i386 libclang-dev llvm-dev\
 automake libgtk-3-dev vtun uml-utilities libxml2 libxml2-dev libboost-all-dev build-essential libtool autoconf unzip wget gcc-8 g++-9 g++-8 -y")
run_command("sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-8 80 --slave /usr/bin/g++ g++ /usr/bin/g++-8")
run_command("sudo update-alternatives --config gcc")
os.chdir(setup_path)

if os.path.exists("ns-3-dev") == False:
    run_command("wget https://gitlab.com/nsnam/ns-3-dev/-/archive/ns-3.38/ns-3-dev-ns-3.38.zip")
    run_command("unzip ns-3-dev-ns-3.38.zip")
    os.remove("ns-3-dev-ns-3.38.zip")
    os.rename("ns-3-dev-ns-3.38","ns-3-dev")
    shutil.copy("../utils/channel/vs-e2e.cc", "ns-3-dev/scratch/")
    os.chdir("ns-3-dev/contrib")
    #os.remove("../src/wifi/test/wifi-emlsr-test.cc")
    run_command("git clone --single-branch --branch 5g-lena-v2.4.y https://gitlab.com/cttc-lena/nr.git nr")
    shutil.copyfile("../../../utils/traffic/xr-traffic-mixer-helper.h", 
                    "nr/utils/traffic-generators/helper/xr-traffic-mixer-helper.h")
    shutil.copyfile("../../../utils/traffic/xr-traffic-mixer-helper.cc", 
                    "nr/utils/traffic-generators/helper/xr-traffic-mixer-helper.cc")
    os.chdir("..")
    run_command("./ns3 configure --enable-examples")
    run_command("./ns3 build")

os.chdir(setup_path)
