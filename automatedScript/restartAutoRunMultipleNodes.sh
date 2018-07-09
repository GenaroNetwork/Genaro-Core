#!/bin/bash

#计数器
i=1
ws=1
#port端口
port=30315

#rpcport端口
rpcport=8549
wsport=8545
cd ../
make geth
cd automatedScript/

./bootnode.sh

sleep 3

if [ ! -f bootnode/bootnode.log ];then
    echo "please run bootnode.sh first"
    exit
fi
bootnode_addr=enode://"$(grep enode bootnode/bootnode.log|tail -n 1|awk -F '://' '{print $2}'|awk -F '@' '{print $1}')""@127.0.0.1:30301"

tmp=`grep enode bootnode/bootnode.log|tail -n 1|awk -F '://' '{print $2}'|awk -F '@' '{print $1}'`
if [ "$tmp" == "" ];then
    echo "node id is empty, please use: bootnode.sh <node_id>";
   	exit
fi

if [ -d "./nohupNodeLog" ];then
	rm -r nohupNodeLog/*
fi

if [ ! -d "./chainNode" ];then
	mkdir ./chainNode
fi

if [ ! -d "./nohupNodeLog" ];then
	mkdir ./nohupNodeLog
fi

#遍历keystore
for line in `cat fileName`
do 
	#kill 端口
	killPort=`lsof -i:$rpcport |awk '{print $2}'|grep -v PID | xargs`
	if [ "$killPort" != "" ];then
    	kill $killPort
	fi
	sleep 1	
	if [ -d "./nohupNodeLog/nohupNode$i.out" ];then
		rm -r ./nohupNodeLog/nohupNode$i.out
	fi
	
	if [ "$ws" -eq "$i" ];then
		killwsPort=`lsof -i:$wsport |awk '{print $2}'|grep -v PID | xargs`
		if [ "$killwsPort" != "" ];then
    		kill $killwsPort
		fi
		sleep 1	
		nohup ./../build/bin/geth --ws  --wsorigins="*" --wsapi "eth,net,web3,admin,personal,miner" --datadir "./chainNode/chainNode$i" --port "$port" --wsport "$wsport" --wsaddr "0.0.0.0"  --bootnodes "$bootnode_addr" --unlock "0x${line##*--}" --password "./password"  --syncmode "full" --mine  > "./nohupNodeLog/nohupNode$i.out" &
		let "i=$i+1"
		let "port=$port+1"
		continue
	fi

	#启动
	nohup ./../build/bin/geth --rpc --rpccorsdomain "*" --rpcvhosts=* --rpcapi "eth,net,web3,admin,personal,miner" --datadir "./chainNode/chainNode$i" --port "$port" --rpcport "$rpcport" --rpcaddr 0.0.0.0  --bootnodes "$bootnode_addr" --unlock "0x${line##*--}" --password "./password"  --syncmode "full" --mine  > "./nohupNodeLog/nohupNode$i.out" &
	let "i=$i+1"
	let "port=$port+1"
	let "rpcport=$rpcport+1"
done
