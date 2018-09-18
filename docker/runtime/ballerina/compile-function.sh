#!/bin/bash
set -e

FUNCTION=${1:?}

mkdir -p /kubeless/func/
cp -r /kubeless/*.bal /kubeless/func/
if [ ! -f /kubeless/kubeless.toml ]; then
    touch /kubeless/kubeless.toml
fi
cp -r /ballerina/files/src/kubeless_run.tpl.bal /kubeless/
sed 's/<<FUNCTION>>/'"${FUNCTION}"'/g' /kubeless/kubeless_run.tpl.bal > /kubeless/kubeless_run.bal
rm /kubeless/kubeless_run.tpl.bal
ballerina build kubeless_run.bal
