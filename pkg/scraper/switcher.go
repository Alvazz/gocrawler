package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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

type ProxyInfo struct {
	IP        string
	Port      int
	Country   string
	Code      string
	Anonymity string
	Google    bool
	SSL       bool
}

type ProxyList []*ProxyInfo

type Switcher struct {
	usedServers   []uint
	currentServer int
	communIP      *IPInfo
	vpnIP         *IPInfo
}

var proxies ProxyList

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

func GetProxyURLs() error {
	transport := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Duration(5*time.Second))
		},
	}
	client := http.Client{
		Transport: &transport,
	}

	req, err := http.NewRequest("GET", "https://free-proxy-list.net", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept-Language", "es-US,es-419;q=0.9,es;q=0.8,en;q=0.7")
	req.Header.Set("DNT", "1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Get("https://free-proxy-list.net")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Println("Response code", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		/* bodyTex, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("Body Response:\n%s\n", string(bodyTex)) */

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return err
		}

		proxies = make(ProxyList, 0)
		doc.Find("table#proxylisttable tbody tr").Each(func(i int, tableRow *goquery.Selection) {
			fmt.Println("Iterando en la fila", i)
			proxyData := tableRow.ChildrenFiltered("td").Map(func(j int, tableCol *goquery.Selection) string {
				fmt.Println("Columna", j)
				text := tableCol.Text()
				fmt.Printf("text: %q\n", text)
				return text
			})
			p, _ := strconv.Atoi(proxyData[1])
			google := strings.ToLower(proxyData[5]) == "yes"
			useSecure := strings.ToLower(proxyData[6]) == "yes"
			prox := &ProxyInfo{
				IP:        proxyData[0],
				Port:      p,
				Code:      proxyData[2],
				Country:   proxyData[3],
				Anonymity: proxyData[4],
				Google:    google,
				SSL:       useSecure,
			}
			proxies = append(proxies, prox)
			fmt.Printf("ProxyInfo: %v\n", prox)
		})
	} else {
		return errors.New("La petición para obtener los proxies revolvio un status code erroneo")
	}
	return nil
}

func (list ProxyList) GetURLs() []string {
	fmt.Println("ProxyList len:", len(list))
	urls := make([]string, len(list))
	for i, prox := range list {
		protocol := "http"
		if prox.SSL {
			protocol += "s"
		}
		urls[i] = fmt.Sprintf("%s://%s:%d", protocol, prox.IP, prox.Port)
	}
	return urls
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
