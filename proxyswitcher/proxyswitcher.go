package proxyswitcher

import (
	"log"

	"golang.org/x/net/proxy"

	"github.com/cretz/bine/control"
)

// proxyswitcher es la estructura encargada de comunicarse con TOR
type proxyswitcher struct{}

// New es la función encargada de instanciar un objeto de proxyswitcher
func New() *proxyswitcher {
	return &proxyswitcher{}
}

// OpenUrl es la encargada de no se que
func (c *proxyswitcher) OpenURL() error {
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:8118", nil, nil)
	if err != nil {
		log.Printf(err)
		return err
	}

	conn, err := dialer.Dial("tcp", "stackoverflow.com:80")
	if err != nil {
		log.Printf(err)
		return err
	}
	defer conn.Close()

	log.Println("conn.LocalAddr string -->", conn.LocalAddr().String())
	log.Println("conn.LocalAddr newtwork -->", conn.LocalAddr().Network())
	log.Println("conn.RemoteAddr string -->", conn.RemoteAddr().String())
	log.Println("conn.RemoteAddr network -->", conn.RemoteAddr().Network())

	return nil
}

// ReNewConnection es la función encargada de que TOR genere una nueva IP
func (c *proxyswitcher) ReNewConnection() error {
	con := control.NewConn(nil)

	err := con.Authenticate("100%tor_go")
	if err != nil {
		log.Println("Ocurrio un error al autenticar con TOR")
		return err
	}

	err = con.Signal("NEWNYM")
	if err != nil {
		log.Println("Ocurrio un error al establecer la señal")
		return err
	}

	err = con.Close()
	if err != nil {
		log.Println("Ocurrio un error al cerrar la conexión con TOR")
		return err
	}

	return nil
}
