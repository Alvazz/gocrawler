package switcher

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sync"
)

func connectToVPN(pidFile string) {
	openVPNCmd := fmt.Sprintf("sudo openvpn --config ~/Downloads/us_california-aes-128-cbc-tcp-dns.ovpn --management localhost 7505 --writepid %s", pidFile)
	cmd := exec.Command("/bin/sh", "-c", openVPNCmd)

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Conexión iniciada con OpenVPN")

	err = cmd.Wait()
	if err != nil {
		log.Println("El programa no termino")
		log.Println(err)
	}
	log.Printf("fin connectToVPN")
}

func checkIP() {
	count := 0
	currentIP, err := exec.Command("curl", "ipinfo.io/ip").Output()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("IP is: %s\n", currentIP)
	for {
		newIP, err := exec.Command("curl", "ipinfo.io/ip").Output()
		if err != nil {
			log.Fatal()
		}
		log.Printf("IP is: %s\n", newIP)
		if string(newIP) != string(currentIP) {
			log.Println("La IP cambio. Descontectando...")
			break
		}
		count++
	}
	log.Println("Iteraciones para conectar:", count)
}

func kill(pidFile string) {
	var pid string
	file, err := os.Open(pidFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		pid = scanner.Text()
	}
	log.Printf("OpenVPN pid: %s", pid)
	killCmd := fmt.Sprintf("sudo kill -15 %s", pid)
	killProcess, err := exec.Command("/bin/sh", "-c", killCmd).Output()
	if err != nil {
		log.Fatalf("Error al enviar la señal: %v", err)
	}
	log.Printf("kill out: %s", string(killProcess))
}

func SwitchIP() {
	var wg sync.WaitGroup
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatal("La información no pudo ser recuperada")
	}
	filepath := path.Join(path.Dir(filename), "./switcher/config/openvpn_pid.txt")

	wg.Add(1)
	go func() {
		connectToVPN(filepath)
		wg.Done()
	}()

	checkIP()
	log.Println("cerrando proceso de OpenVPN")
	kill(filepath)

	wg.Wait()
	log.Println("Desconectado de la VPN")
}

func TestConnection(ctx context.Context, cancel context.CancelFunc) {
	var wg sync.WaitGroup

	cmd := exec.CommandContext(ctx, "sleep", "10")
	wg.Add(1)
	go func() {
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Conexión iniciada")

		err = cmd.Wait()
		if err != nil {
			log.Println("El programa no termino")
			log.Println(err)
		}
		log.Printf("fin de la conexión")
		wg.Done()
	}()
	wg.Wait()
	log.Printf("proceso de conexión terminado")
}
