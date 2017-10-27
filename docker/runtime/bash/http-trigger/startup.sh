#!/bin/bash

if [[ -f /kubeless/requirements ]]; then
 install_packages -y $(cat /kubeless/requirements)
fi

echo "Starting up runtime for bash functions"
python /kubeless-shell.py 
