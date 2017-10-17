#!/bin/bash

if [[ -f /tmp/requirements ]]; then
 install_packages -y $(cat /tmp/requirements)
fi

echo "yes, I have run it!!!!"
