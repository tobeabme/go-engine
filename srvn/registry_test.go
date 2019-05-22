package srvn

import (
	"fmt"
	"testing"
)

//go test -v -run=TestRegistryToConsul
func TestRegistryToConsul(t *testing.T) {
	endpoints := []string{"127.0.0.1:8500"}

	// var cfg Config
	cfg := Config{Provider: Consul, Endpoints: endpoints}
	fmt.Println(cfg)

	s, err := NewRegister(cfg)
	if err != nil {
		panic(err)
	}

	err = s.Register("go_service_passport", "127.0.0.1:50051")
	if err == nil {
		fmt.Println("It's successful to register to consul.")
	}
}

//go test -v -run=TestRegistryToEtcd
func TestRegistryToEtcd(t *testing.T) {

}
