// Adapted from David Knezić https://github.com/davidknezic/desk/blob/master/bridge.go
// Licensed under MIT (c) David Knezić (https://github.com/davidknezic/desk/blob/master/LICENSE)

package main

import (
	"net/http"
	"github.com/tarm/serial"
	"log"
	"strconv"
	"time"
)

const UP_HEIGHT = 90
const DOWN_HEIGHT = 10

type DeskHandler struct {
	outgoing chan Message
	incoming chan Message
}

type MessageType byte

const (
	// checking the availability of the desk
	TypeAliveRequest  MessageType = 0x01
	TypeAliveResponse MessageType = 0x02

	// setting the height of the desk
	TypeSetHeightRequest MessageType = 0x03

	// querying the height of the desk
	TypeGetHeightRequest  MessageType = 0x04
	TypeGetHeightResponse MessageType = 0x05

	// stopping the desk
	TypeStopRequest MessageType = 0x06

	// TODO: to be implemented
	TypeGetStatusRequest  MessageType = 0x07
	TypeGetStatusResponse MessageType = 0x08

	// moving the desk
	TypeMoveUpRequest   MessageType = 0x0A
	TypeMoveDownRequest MessageType = 0x0B

	// the desk notifying about a height change
	TypeUpdateHeightEvent MessageType = 0x0C
)

type Message struct {
	Type  MessageType
	Value byte
}

func receiver(c chan<- Message, p *serial.Port) {
	message := make([]byte, 3)
	buf := make([]byte, 1)

	for {
		// shift bytes to left
		message[0] = message[1]
		message[1] = message[2]

		// read new byte
		_, err := p.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		// append new byte
		message[2] = buf[0]

		// checksum
		if message[0]+message[1] != message[2] {
			continue
		}

		c <- Message{Type: MessageType(message[0]), Value: message[1]}
	}
}

func sender(c <-chan Message, p *serial.Port) {
	message := make([]byte, 3)

	for {
		// get a message
		m := <-c

		log.Println("Message to send", m)

		// fill the message buffer
		message[0] = byte(m.Type)
		message[1] = m.Value

		// calculate the checksum
		message[2] = message[0] + message[1]

		log.Println("Buffer to send", message)

		// write the buffer
		_, err := p.Write(message)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func heightToPercentage(height int) int {
	return (height - 68) / 50
}

func heightPercentageToCentimeters(percentage int) int {
	factor := float64(percentage) / 100.0
	return int(68.0 + 50.0*factor)
}

func main() {

	log.Println("Opening serial console...")
	c := &serial.Config{
		Name: "/dev/ttyS0",
		Baud: 9600,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Println("Failed opening serial console.")
		log.Fatal(err)
	}

	var outgoing chan Message = make(chan Message)
	var incoming chan Message = make(chan Message)
	go sender(outgoing, s)
	go receiver(incoming, s)

	log.Println("Starting desk-server...")
	// http.HandleFunc("/", DeskControlServer)
	myFilesHandler := &DeskHandler{outgoing: outgoing, incoming: incoming}
    http.HandleFunc("/", myFilesHandler.handler)

    http.ListenAndServe(":9987", nil)
}

func (dh *DeskHandler) handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "down"{
		log.Println("Received DOWN request")
		setPosition(DOWN_HEIGHT, dh.incoming, dh.outgoing)
	} else if r.URL.Path[1:] == "up"{
		log.Println("Received UP request")
		setPosition(UP_HEIGHT, dh.incoming, dh.outgoing)
	} else if r.URL.Path[1:] == "toggle"{
		log.Println("Received toggle request")
		go togglePosition(dh.incoming, dh.outgoing)		
	} else{
		log.Println("Received specific set height request")
		x, err := strconv.Atoi(r.URL.Path[1:])
		if err == nil && x > 0{
			log.Println(w, "Setting height to %s", x)
			setPosition(x, dh.incoming, dh.outgoing)
		} else{
			log.Println(w, "Error parsing height")
		}
	}
}


func togglePosition(incoming chan Message, outgoing chan Message){
	time.Sleep(2000 * time.Millisecond)
	outgoing <- Message{Type: TypeGetHeightRequest}
	hi := <-incoming
	log.Println("Height is", hi.Value)
	int_hi := int(hi.Value)
	if int_hi >= 100 {
		log.Println("Toggle DOWN")
		setPosition(DOWN_HEIGHT, incoming, outgoing)
	} else if int_hi < 100 && int_hi >= 1 {
		log.Println("Toggle UP")
		setPosition(UP_HEIGHT, incoming, outgoing)
	}
}

func setPosition(position int, incoming chan Message, outgoing chan Message){
	log.Println("Setting desk to", position, "percent height")
	height := heightPercentageToCentimeters(position)
	log.Println("This corresponds to", height, "cm height")
	outgoing <- Message{Type: MessageType(TypeSetHeightRequest), Value: byte(height)}
}

