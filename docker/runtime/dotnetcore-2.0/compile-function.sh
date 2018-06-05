#!/bin/bash

PROJECT_MOUNT=$1
PACKAGES_DIR=$PROJECT_MOUNT/packages
USER_CSPROJ=$PROJECT_MOUNT/project.csproj
DEFAULT_CSPROJ=/app/project.csproj 

if [ -s $USER_CSPROJ ]; then
	echo .csproj present;
else
	cp $DEFAULT_CSPROJ $USER_CSPROJ
fi

dotnet restore $PROJECT_MOUNT --packages $PACKAGES_DIR
dotnet publish $PROJECT_MOUNT -o publish -c Release --no-restore