#!/bin/bash

declare -a arr=(16 32 64 128 256 512 1024 2048)

arrayLength=${#arr[@]i}-1

for (( i=0; i<${arrayLength}; i++))
do
echo "$1 $arr[$i] $arrayLength"
sips -z ${arr[$i]} ${arr[$i]} $1 --out icons.iconset/icon_${arr[$i]}x${arr[$i]}.png
sips -z ${arr[$i+1]} ${arr[$i+1]} $1 --out icons.iconset/icon_${arr[$i]}x${arr[$i]}@2x.png
done

iconutil -c icns -o icon.icns icons.iconset
