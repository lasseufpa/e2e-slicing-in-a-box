#!/bin/bash

echo "Installing NS-3 dependencies"
sudo apt install g++ python3 cmake ninja-build git -y 
sudo apt install ccache -y 
sudo apt install g++ python3 -y  
sudo apt install gir1.2-goocanvas-2.0 python3-gi python3-gi-cairo python3-pygraphviz gir1.2-gtk-3.0 ipython3 -y 
sudo apt install python3-setuptools git -y 
sudo apt install g++ python3 python3-dev pkg-config sqlite3 cmake -y 
sudo apt install qtbase5-dev qtchooser qt5-qmake qtbase5-dev-tools -y 
sudo apt install qt5-default -y 
sudo apt install gir1.2-goocanvas-2.0 python3-gi python3-gi-cairo python3-pygraphviz gir1.2-gtk-3.0 ipython3 -y 
sudo apt install python-pygraphviz python-kiwi python-pygoocanvas libgoocanvas-dev ipython -y 
sudo apt install openmpi-bin openmpi-common openmpi-doc libopenmpi-dev -y 
sudo apt install mercurial unzip -y 
sudo apt install gdb valgrind  -y 
sudo apt install clang-format -y 
sudo apt install doxygen graphviz imagemagick -y 
sudo apt install texlive texlive-extra-utils texlive-latex-extra texlive-font-utils dvipng latexmk -y 
sudo apt install python3-sphinx dia  -y 
sudo apt install gsl-bin libgsl-dev libgslcblas0 -y 
sudo apt install tcpdump -y 
sudo apt install sqlite sqlite3 libsqlite3-dev -y 
sudo apt install libxml2 libxml2-dev -y 
sudo apt install cmake libc6-dev libc6-dev-i386 libclang-dev llvm-dev automake python3-pip -y 
sudo apt install libgtk-3-dev -y 
sudo apt install vtun lxc uml-utilities -y 
sudo apt install libxml2 libxml2-dev libboost-all-dev -y 
sudo pip install cppyy
sudo pip install distro 

echo "Building NS-38 and installing:"
wget https://www.nsnam.org/releases/ns-allinone-3.38.tar.bz2
tar xfj ns-allinone-3.38.tar.bz2
cd ns-allinone-3.38/ns-3.38
./ns3 configure --enable-examples --enable-tests
./ns3 build
