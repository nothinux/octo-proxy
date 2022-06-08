package proxy

import (
	"log"

	"github.com/nothinux/octo-proxy/pkg/config"
)

func SendData(hc config.HostConfig, message []byte) error {
	d, err := dialTarget(hc)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = d.Write(message)
	if err != nil {
		log.Println(err)
		return err
	}
	d.Close()

	return nil
}
