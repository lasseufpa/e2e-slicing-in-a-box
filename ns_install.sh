#!/bin/bash

echo "Installing NS-3 dependencies"
sudo apt update
sudo apt install ccache gir1.2-goocanvas-2.0 python3-gi python3-gi-cairo python3-pygraphviz gir1.2-gtk-3.0 ipython3 python3-setuptools\
 qtbase5-dev qtchooser qt5-qmake qtbase5-dev-tools qt5-default openmpi-bin openmpi-common openmpi-doc libopenmpi-dev mercurial gdb valgrind\
 clang-format gsl-bin libgsl-dev libgslcblas0 tcpdump sqlite sqlite3 libsqlite3-dev cmake libc6-dev libc6-dev-i386 libclang-dev llvm-dev\
 automake libgtk-3-dev vtun lxc uml-utilities libxml2 libxml2-dev libboost-all-dev build-essential libtool autoconf unzip wget gcc-8 g++-8 -y 

sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-8 80 --slave /usr/bin/g++ g++ /usr/bin/g++-8 --slave /usr/bin/gcov gcov /usr/bin/gcov-8

sudo update-alternatives --config gcc

echo "Building NS-38 and installing:"
wget https://www.nsnam.org/releases/ns-allinone-3.38.tar.bz2
tar xfj ns-allinone-3.38.tar.bz2
cd ns-allinone-3.38/ns-3.38
./ns3 configure --enable-examples --enable-tests
./ns3 build
