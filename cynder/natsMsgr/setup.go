package natsMsgr

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"os"
)

func ConnectToNats() *nats.Conn {

	// Connect to natsMsgr server
	username := os.Getenv("NATS_USERNAME")
	password := os.Getenv("NATS_PASSWORD")
	hostname := os.Getenv("NATS_HOSTNAME")
	port := os.Getenv("NATS_PORT")

	url := fmt.Sprintf("nats://%s:%s@%s:%s", username, password, hostname, port)
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatalf("Error connecting to nats: %v \n\nURL: %s", err, url)
	}
	//defer nc.Close()
	log.Println("Connected to nats!")

	return nc
}
