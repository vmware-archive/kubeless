#!/bin/sh
echo "Running init.sh script"

if [ -f /kubeless/requirements ]; then
	echo "Installing requirements inside the runtime"
	install_packages -y $(cat /kubeless/requirements)
fi
echo "Starting up runtime for bash functions" 


exec python /kubeless-shell.py  

