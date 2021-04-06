#!/bin/bash

VPN_CONFIGFILES=("sweden" "argentina" "uk_london" "spain" "uk_southampton" "us_california" "mexico" "us_silicon_valley" "us_east" "us_west" "us_texas")

test_connection() {
	local IP_INFO=$(curl -s -m 5 ipinfo.io/json)
	while [ $? -ne 0 ] || [ -z "$IP_INFO" ]; do
		echo "Offline"
		IP_INFO=$(curl -s -m 5 ipinfo.io/json)
	done
	echo "Online"
	echo $IP_INFO
	local IP=$(echo "$IP_INFO" | jq --raw-output '.ip')
	eval "$1=$IP"
}

is_ip_change() {
	local IP_VPN=""
	test_connection IP_VPN
	echo "Nueva IP: $IP_VPN | IP Anterior: $DEFAULT_IP "
	if [ $1 = "-eq" ]; then
		while [ "$DEFAULT_IP" = "$IP_VPN" ]; do
			echo "Obteniendo IP nuevamente..."
			sleep 2
			test_connection IP_VPN
			echo "Nueva IP: $IP_VPN | IP Anterior: $DEFAULT_IP "
		done
	else
		while [ "$DEFAULT_IP" != "$IP_VPN" ]; do
			echo "Obteniendo IP nuevamente..."
			sleep 2
			test_connection IP_VPN
			echo "Nueva IP: $IP_VPN | IP Anterior: $DEFAULT_IP "
		done
	fi
}

change_connection_status() {
	local status=0
	case $2 in
		"start") systemctl start pia@$1;;
		"stop") systemctl stop pia@$1;;
		*) status=1 ;;
	esac
	return $status
}

rotate_ip() {
	local len=${#VPN_CONFIGFILES[@]}
	a=$( shuf -i 0-$(expr $len - 1))
	for i in ${a[@]}; do
		local CURRENT_CONFIG=${VPN_CONFIGFILES[$i]}
		echo "$i $CURRENT_CONFIG"

		change_connection_status $CURRENT_CONFIG start
		echo $?
		if [ $? -ne 0 ]; then
			echo "Falló el servicio de la VPN"
		else
			is_ip_change -eq
			echo "Conectado a $CURRENT_CONFIG"
		fi
		sleep 2

		crawler
		#crawler &> $HOME/crawling-data/logs/$(date +"%Y%m%d_%H%M%S").log
		
		change_connection_status $CURRENT_CONFIG stop
		echo $?
		if [ $? -ne 0 ]; then
			echo "Falló el servicio de la VPN"
		else
			is_ip_change -ne
			echo "Desconectado de $CURRENT_CONFIG"
		fi
		sleep 2
		break
	done
}

if [ -z ${PROJECTPATH} ]; then
	export PROJECTPATH=$(dirname ${BASH_SOURCE[0]})
fi

DEFAULT_IP=""
test_connection DEFAULT_IP
echo "IP DEFAULT: $DEFAULT_IP"
rotate_ip

