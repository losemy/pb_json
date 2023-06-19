#!/bin/bash
# encoding:utf-8
#stop process
pidInfo=$(ps -ef | grep "pb_json" | grep -v grep | awk '{print $2}')
echo "`date` old pid info is $pidInfo"

for pid in $pidInfo; do
    kill -9 $pid
done

sleep 3

source start.sh

pidInfo=$(ps -ef | grep "pb_json" | grep -v grep | awk '{print $2}')
echo "`date` new pid info is $pidInfo"