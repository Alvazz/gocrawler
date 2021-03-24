package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os/exec"
	"sort"
	"time"
)

var vpnServersPool = []string{
	"uk_manchester",
	"uk_london",
	"uk_southampton",
	"us_california",
	"us_silicon_valley",
	"us_east",
	"us_west",
}

type IPInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Timezone string `json:"timezone"`
}

type Switcher struct {
	usedServers   []uint
	currentServer int
	communIP      *IPInfo
	vpnIP         *IPInfo
}

func NewSwitcher() *Switcher {
	var ipinfo *IPInfo
	var online bool
	for {
		ipinfo, online = IsOnline()
		if online {
			break
		}
	}

	return &Switcher{
		usedServers:   []uint{},
		currentServer: -1,
		communIP:      ipinfo,
	}
}

func (sw *Switcher) GetCurrentServer() int {
	return sw.currentServer
}

func (sw *Switcher) GetUsedServers() []uint {
	return sw.usedServers
}

func (sw *Switcher) GetCommunIP() *IPInfo {
	return sw.communIP
}

func (sw *Switcher) GetVpnIP() *IPInfo {
	return sw.vpnIP
}

func (sw *Switcher) Connect() error {
	i := -1
	fmt.Printf("len(usedServers): %v\n", len(sw.usedServers))
	fmt.Printf("len(vpnServersPool): %v\n", len(vpnServersPool))

	if len(sw.usedServers) == len(vpnServersPool) {
		log.Println("Reiniciando pool de conexiones")
		sw.usedServers = make([]uint, 0)
	}
	for i < len(sw.usedServers) {
		sw.currentServer = rand.Intn(len(vpnServersPool))
		fmt.Printf("currentServer: %d\n", sw.currentServer)
		i = sort.Search(len(sw.usedServers), func(idx int) bool {
			fmt.Printf("sw.usedServers[idx]: %v\n", sw.usedServers[idx])
			return uint(sw.currentServer) == sw.usedServers[idx]
		})
		fmt.Printf("encontrado: %d\n", i)
	}

	sw.usedServers = append(sw.usedServers, uint(sw.currentServer))
	configName := vpnServersPool[sw.currentServer]

	fmt.Printf("usedServers: %v\n", sw.usedServers)
	fmt.Printf("configName: %v\n", configName)

	service := fmt.Sprintf("pia@%s", configName)
	cmd := exec.Command("systemctl", "start", service)
	log.Printf("Iniciando conexión VPN a %s", configName)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (sw *Switcher) Disconnect() error {
	if sw.currentServer == -1 {
		return nil
	}
	configName := vpnServersPool[sw.currentServer]
	service := fmt.Sprintf("pia@%s", configName)
	cmd := exec.Command("systemctl", "stop", service)
	log.Printf("Deteniendo conexión VPN a %s", configName)
	err := cmd.Run()
	if err != nil {
		return err
	}
	sw.currentServer = -1
	return nil
}

func IsOnline() (*IPInfo, bool) {
	transport := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Duration(5*time.Second))
		},
	}
	client := http.Client{
		Transport: &transport,
	}

	res, err := client.Get("https://ipinfo.io/json")
	if err != nil {
		log.Println("ERROR HTTP GET", err)
		return nil, false
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, true
	}

	defaultIP := IPInfo{}
	jsonErr := json.Unmarshal(body, &defaultIP)
	if jsonErr != nil {
		return nil, true
	}
	log.Printf("IPResponse: %+v", defaultIP)
	return &defaultIP, true
}

func (sw *Switcher) RotateIP() error {
	if sw.currentServer != -1 {
		if err := sw.Disconnect(); err != nil {
			return fmt.Errorf("Error al desconectar la VPN: %v", err)
		}
		log.Println("Sin error al desconectar")
		disconnectCount := 0
		for {
			if disconnectCount > 15 {
				log.Println("Volviendo a desconectar")
				if err := sw.Disconnect(); err != nil {
					return fmt.Errorf("Error al desconectar la VPN: %v", err)
				}
			}
			log.Println("Esperando...")
			time.Sleep(time.Duration(2 * time.Second))
			log.Println("Comprando conexión e IP")
			ipinfo, online := IsOnline()
			if online && ipinfo.IP == sw.GetCommunIP().IP {
				log.Printf("IP default: %+v", ipinfo)
				log.Printf("ipinfo.IP: %s\tCommunIP: %s", ipinfo.IP, sw.GetCommunIP().IP)
				break
			}
			log.Println("Aún conectado a la VPN")
		}
		log.Println("IP Restablecida")
	}

	if err := sw.Connect(); err != nil {
		return fmt.Errorf("Connect error: %v", err)
	}
	log.Println("Sin error al conectar")
	connectCount := 0
	for {
		if connectCount > 15 {
			log.Println("Volviendo a conectar")
			if err := sw.Connect(); err != nil {
				return fmt.Errorf("Connect error: %v", err)
			}
		}
		log.Println("Esperando...")
		time.Sleep(time.Duration(2 * time.Second))
		log.Println("Comprando conexión e IP")
		ipinfo, online := IsOnline()
		if online && ipinfo.IP != sw.GetCommunIP().IP {
			log.Printf("Nueva IP: %+v", ipinfo)
			break
		}
		log.Println("Sin nueva IP")
	}
	log.Println("Nueva IP obtenida con éxito")

	return nil
}
