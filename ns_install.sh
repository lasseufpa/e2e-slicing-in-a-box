#!/bin/bash

if [[ $EUID -ne 0 ]]; then
   echo "Must bee executed as sudo!"
   echo "Error 1"
   exit 1
fi
echo "Installing NS-3 dependencies"
apt install g++ python3 cmake ninja-build git -y 
apt install ccache -y 
apt install g++ python3 -y 
python3 -m pip install --user cppyy -y 
apt install gir1.2-goocanvas-2.0 python3-gi python3-gi-cairo python3-pygraphviz gir1.2-gtk-3.0 ipython3 -y 
apt install python3-setuptools git -y 
apt install g++ python3 python3-dev pkg-config sqlite3 cmake -y 
apt install qtbase5-dev qtchooser qt5-qmake qtbase5-dev-tools -y 
apt install qt5-default -y 
apt install gir1.2-goocanvas-2.0 python3-gi python3-gi-cairo python3-pygraphviz gir1.2-gtk-3.0 ipython3 -y 
apt install python-pygraphviz python-kiwi python-pygoocanvas libgoocanvas-dev ipython -y 
apt install openmpi-bin openmpi-common openmpi-doc libopenmpi-dev -y 
apt install mercurial unzip -y 
apt install gdb valgrind  -y 
apt install clang-format -y 
apt install doxygen graphviz imagemagick -y 
apt install texlive texlive-extra-utils texlive-latex-extra texlive-font-utils dvipng latexmk -y 
apt install python3-sphinx dia  -y 
apt install gsl-bin libgsl-dev libgslcblas0 -y 
apt install tcpdump -y 
apt install sqlite sqlite3 libsqlite3-dev -y 
apt install libxml2 libxml2-dev -y 
apt install cmake libc6-dev libc6-dev-i386 libclang-dev llvm-dev automake python3-pip -y 
python3 -m pip install --user cxxfilt -y 
apt install libgtk-3-dev -y 
apt install vtun lxc uml-utilities -y 
apt install libxml2 libxml2-dev libboost-all-dev -y 
pip install cppyy
pip install distro 

echo "Building NS-38 and installing:"
git clone https://gitlab.com/nsnam/bake
export BAKE_HOME=`pwd`/bake 
export PATH=$PATH:$BAKE_HOME
export PYTHONPATH=$PYTHONPATH:$BAKE_HOME
bake.py check
bake.py configure -e ns-3.38
bake.py show   
bake.py deploy
