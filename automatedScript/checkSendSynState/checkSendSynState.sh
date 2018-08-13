#!/bin/bash
time=`date +%s`
pkill SendSynState
sleep 3
nohup ./../../cmd/SendSynState/SendSynState -u http://127.0.0.1:9003 -t 1 -a 0xebb97ad3ca6b4f609da161c0b2b0eaa4ad58f3e8 > ./checkSendSynStateLogInfo/${time}SendSynState.log 2>&1 &
