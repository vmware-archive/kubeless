#!/bin/bash

set -e

/usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf &
/proxy
