#!/bin/bash

#-------------------------------------------------------------------------------
###LICENCE
    ##Copyright (C) 2023, Ewo On sight, GPLv3+
    ##Maintainer: Ewo On Sight,
    ##Contact: <thetime_traywick@8shield.net>

        #This program is free software: you can redistribute it and/or modify
        #it under the terms of the GNU General Public License as published by
        #the Free Software Foundation, either version 3 of the GPL License, or
        #any later version.

        #This program is distributed in the hope that it will be useful,
        #but WITHOUT ANY WARRANTY; without even the implied warranty of
        #MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
        #GNU General Public License for more details.

        #You should have received a copy of the GNU General Public License
        #along with this program.  If not, see <https://www.gnu.org/licenses/.

#--------------------------------------------------------------------------------

# This is niceland-update v0.1


# this file must be present to track if the release has changed
RELEASE_FILE="/home/anon/release"
 
# get the release tag from the latest release URL
RELEASE_TAG=$(curl -Ls -o /dev/null -w %{url_effective} https://github.com/tunnels-is/nicelandVPN/releases/latest | awk -F '/' '{print $NF}')
 
# Check if the release has changed
if [ -f "$RELEASE_FILE" ] && [ "$(cat "$RELEASE_FILE")" = "$RELEASE_TAG" ]; then
    echo "Release has not changed"
else
    echo "Release has changed"
 
    # Replace the 'v' prefix in the release tag with a '-' character
    DOWNLOAD_TAG=$(echo "$RELEASE_TAG" | sed 's/^v/-/')
 
    DOWNLOAD_URL="https://github.com/tunnels-is/nicelandVPN/releases/download/$RELEASE_TAG/nicelandVPN$DOWNLOAD_TAG-linux.pacman"
    wget $DOWNLOAD_URL
 
    #echo "nicelandVPN$DOWNLOAD_TAG-linux.pacman"
    sudo pacman -U "nicelandVPN$DOWNLOAD_TAG-linux.pacman"
 
    # Update the release file with the new release tag
    echo "$RELEASE_TAG" > "$RELEASE_FILE"
    
    # Removes downloaded package after install
    rm -rf ./nicelandVPN-*-linux.*

    # Notifies that the update has been completed
    printf "\nUpdate completed. \nnicelandVPN has been updated to the latest version\n"
fi