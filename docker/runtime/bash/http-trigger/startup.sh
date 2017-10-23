#!/bin/bash

if [[ -f /kubeless/requirements ]]; then
 install_packages -y $(cat /kubeless/requirements)
fi

echo "yes, I have run it!!!!"
python /kubeless-shell.py 
