#!/bin/bash

VPN_CONFIGFILES=("sweden" "uk_manchester" "uk_london" "spain" "uk_southampton" "us_california" "mexico" "us_silicon_valley" "us_east" "us_west")

test_connection() {
	local IP_INFO=$(curl -s -m 5 ipinfo.io/json)
	while [ $? -ne 0 ]; do
		echo "Offline"
		IP_INFO=$(curl -s -m 5 ipinfo.io/json)
	done
	echo "Online"
	echo $IP_INFO
	local IP=$(echo "$IP_INFO" | jq --raw-output '.ip')
	eval "$1=$IP"
}

connect_to_vpn() {
	local len=${#VPN_CONFIGFILES[@]}
	a=$( shuf -i 0-$(expr $len - 1))
	for i in ${a[@]}; do
		local CURRENT_CONFIG=${VPN_CONFIGFILES[$i]}
		echo "$i $CURRENT_CONFIG"
		systemctl start pia@$CURRENT_CONFIG
		local IP_VPN=""
		test_connection IP_VPN
		echo "Nueva IP: $IP_VPN | IP Anterior: $DEFAULT_IP "
		while [ "$DEFAULT_IP" = "$IP_VPN" ]; do
			echo "Probando nuevamente..."
			sleep 2
			test_connection IP_VPN
			echo $IP_VPN
		done
        echo "Conectado a $CURRENT_CONFIG"
        sleep 3
		clear
		crawler

		systemctl stop pia@$CURRENT_CONFIG
		test_connection IP_VPN
		echo $IP_VPN
		while [ "$DEFAULT_IP" != "$IP_VPN" ]; do
			echo "Desconectando nuevamente..."
			sleep 2
			test_connection IP_VPN
			echo $IP_VPN
		done
        echo "Desconectado de $CURRENT_CONFIG"
        sleep 3
		clear
	done
}

if [ -z ${PROJECTPATH} ]; then
  export PROJECTPATH=$(dirname ${BASH_SOURCE[0]})
fi

DEFAULT_IP=""
test_connection DEFAULT_IP
echo "IP DEFAULT: $DEFAULT_IP"
connect_to_vpn
echo "IP_VPN: $IP_VPN IP: $IP"

