#!/bin/bash

if [[ $EUID -ne 0 ]]; then
   echo "Must bee executed as sudo!"
   echo "Error 1"
   exit 1
fi
echo "Installing NS-3 dependencies"
echo -n "["
while true; do echo -n "."; sleep 5; done &
trap "kill $!" SIGTERM SIGKILL
apt install g++ python3 cmake ninja-build git &>/dev/null 
apt install ccache &>/dev/null 
apt install g++ python3 &>/dev/null 
python3 -m pip install --user cppyy &>/dev/null 
apt install gir1.2-goocanvas-2.0 python3-gi python3-gi-cairo python3-pygraphviz gir1.2-gtk-3.0 ipython3 &>/dev/null 
apt install python3-setuptools git &>/dev/null 
apt install g++ python3 python3-dev pkg-config sqlite3 cmake &>/dev/null 
apt install qtbase5-dev qtchooser qt5-qmake qtbase5-dev-tools &>/dev/null 
apt install qt5-default &>/dev/null 
apt install gir1.2-goocanvas-2.0 python3-gi python3-gi-cairo python3-pygraphviz gir1.2-gtk-3.0 ipython3 &>/dev/null 
apt install python-pygraphviz python-kiwi python-pygoocanvas libgoocanvas-dev ipython &>/dev/null 
apt install openmpi-bin openmpi-common openmpi-doc libopenmpi-dev &>/dev/null 
apt install mercurial unzip &>/dev/null &
apt install gdb valgrind  &>/dev/null 
apt install clang-format &>/dev/null 
apt install doxygen graphviz imagemagick &>/dev/null 
apt install texlive texlive-extra-utils texlive-latex-extra texlive-font-utils dvipng latexmk &>/dev/null 
apt install python3-sphinx dia  &>/dev/null 
apt install gsl-bin libgsl-dev libgslcblas0 &>/dev/null 
apt install tcpdump &>/dev/null 
apt install sqlite sqlite3 libsqlite3-dev &>/dev/null 
apt install libxml2 libxml2-dev &>/dev/null 
apt install cmake libc6-dev libc6-dev-i386 libclang-dev llvm-dev automake python3-pip &>/dev/null 
python3 -m pip install --user cxxfilt &>/dev/null 
apt install libgtk-3-dev &>/dev/null 
apt install vtun lxc uml-utilities &>/dev/null 
apt install libxml2 libxml2-dev libboost-all-dev &>/dev/null 
echo -n "]"
echo
echo "Concluded."

echo "Building NS-38 and installing:"
git clone https://gitlab.com/nsnam/bake
export BAKE_HOME=`pwd`/bake 
export PATH=$PATH:$BAKE_HOME
export PYTHONPATH=$PYTHONPATH:$BAKE_HOME
bake.py check
bake.py configure -e ns-3.38
bake.py show   
bake.py deploy

