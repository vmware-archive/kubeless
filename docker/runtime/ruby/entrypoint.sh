#!/bin/bash

set -e

ruby /kubeless.rb &
/proxy
