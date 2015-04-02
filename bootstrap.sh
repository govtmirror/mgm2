#!/usr/bin/env bash

apt-get update
apt-get install -y golang git npm

ln -s /usr/bin/nodejs /usr/bin/node
npm install -g bower
npm install -g grunt
npm install -g yo
npm install -g generator-angular

mkdir /home/vagrant/go
mkdir /home/vagrant/go/bin
mkdir /home/vagrant/go/src

echo 'export GOPATH=$HOME/dev/go' >> /home/vagrant/.bashrc
echo 'export GOBIN=$GOPATH/bin' >> /home/vagrant/.bashrc
echo 'export PATH=$PATH:$GOBIN' >> /home/vagrant/.bashrc

