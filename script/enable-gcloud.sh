#!/bin/bash
set -e
BUILD_DIR=${1:?}
export GOOGLE_APPLICATION_CREDENTIALS=$BUILD_DIR/client_secrets.json
echo $GCLOUD_KEY > $GOOGLE_APPLICATION_CREDENTIALS
if [ ! -d $HOME/gcloud/google-cloud-sdk ]; then
    mkdir -p $HOME/gcloud &&
    wget -q https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-187.0.0-linux-x86_64.tar.gz --directory-prefix=$HOME/gcloud &&
    cd $HOME/gcloud &&
    tar xzf google-cloud-sdk-187.0.0-linux-x86_64.tar.gz &&
    printf '\ny\n\ny\ny\n' | ./google-cloud-sdk/install.sh &&
    sudo ln -s $HOME/gcloud/google-cloud-sdk/bin/gcloud /usr/local/bin/gcloud
    cd $BUILD_DIR;
fi
gcloud -q config set project $GKE_PROJECT
if [ -a $GOOGLE_APPLICATION_CREDENTIALS ]; then
    gcloud -q auth activate-service-account --key-file $GOOGLE_APPLICATION_CREDENTIALS;
fi
