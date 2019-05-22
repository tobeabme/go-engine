package srvn

import (
	"log"
	"net"
)

// endpoint usability detection
func ConsulEndpointCheck(endpoints []string) []string {
	for _, v := range endpoints {
		conn, err := net.Dial("udp", v)
		if err != nil {
			log.Println("EndpointCheck", v, err)
			continue
		}

		defer conn.Close()
		return []string{v}
	}

	return []string{"127.0.0.1:8500"}
}

func EtcdEndpointCheck(endpoints []string) []string {
	for _, v := range endpoints {
		conn, err := net.Dial("udp", v)
		if err != nil {
			log.Println("EndpointCheck", v, err)
			continue
		}

		defer conn.Close()
		return []string{v}
	}

	return []string{"127.0.0.1:2379"}
}
