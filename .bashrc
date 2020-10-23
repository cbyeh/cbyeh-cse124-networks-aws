# Source global definitions
if [ -f /etc/bashrc ]; then
	. /etc/bashrc
fi

# .bashrc
export PATH=$PATH:/usr/local/go/bin     # making sure go is on path
export GOPATH=$HOME/PA2
export PATH=$PATH:$GOPATH/bin

# User specific aliases and functions
