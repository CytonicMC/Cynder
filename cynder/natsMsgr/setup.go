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

	nc, err := nats.Connect(fmt.Sprintf("natsMsgr://%s:%s@%s:%s", username, password, hostname, port))
	if err != nil {
		log.Fatalf("Error connecting to natsMsgr: %v", err)
	}
	//defer nc.Close()
	log.Println("Connected to natsMsgr!")

	return nc
}
