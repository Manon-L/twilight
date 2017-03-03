package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

type state int

const (
	waiting state = iota
	p1
	p2
	ready
)

type server struct {
	*Map
	name [2]string
}

func (s *server) run() {
	log.Println("Starting tcp server")
	l, err := net.Listen("tcp", ":5555")
	defer l.Close()
	if err != nil {
		panic(err.Error())
	}
	// first player
	con1, err := l.Accept()
	if err != nil {
		panic(err.Error())
	}
	log.Println("First player entered the game")
	go s.runP(con1, 0)
	// second player
	con2, err := l.Accept()
	if err != nil {
		panic(err.Error())
	}
	log.Println("second player entered the game")
	go s.runP(con2, 1)
}

func (s *server) set(c net.Conn) {
	msg := make([]byte, 5)
	copy(msg, []byte("SET"))
	msg[3] = byte(uint8(s.Rows))
	msg[4] = byte(uint8(s.Columns))
	c.Write(msg)
}

func (s *server) hum(c net.Conn) {
	n := len(s.humans)
	msg := make([]byte, 4+2*n)
	copy(msg, []byte("HUM"))
	msg[3] = byte(uint8(n))
	for i, h := range s.humans {
		hum := s.cells[h]
		msg[4+2*i] = byte(uint8(hum.x))
		msg[4+2*i+1] = byte(uint8(hum.y))
	}
	c.Write(msg)
}

func (s *server) hme(c net.Conn, id int) {
	msg := make([]byte, 5)
	copy(msg, []byte("HME"))
	mon := s.cells[s.monster[id][0]]
	msg[3] = byte(uint8(mon.x))
	msg[4] = byte(uint8(mon.y))
	c.Write(msg)
}

func (s *server) state(c net.Conn, trame string) {
	n := len(s.humans) + len(s.monster[0]) + len(s.monster[1])
	msg := make([]byte, 4+5*n)
	copy(msg, []byte(trame))
	msg[3] = byte(uint8(n))
	var i int
	// Send all humans data
	for _, h := range s.humans {
		hum := s.cells[h]
		msg[4+2*i] = byte(uint8(hum.x))
		msg[4+2*i+1] = byte(uint8(hum.y))
		msg[4+2*i+2] = byte(uint8(hum.count))
		i++
	}
	// Send all vamp data
	for _, m := range s.monster[1] {
		mon := s.cells[m]
		msg[4+2*i] = byte(uint8(mon.x))
		msg[4+2*i+1] = byte(uint8(mon.y))
		msg[4+2*i+3] = byte(uint8(mon.count))
		i++
	}
	// Send all wolf data
	for _, m := range s.monster[0] {
		mon := s.cells[m]
		msg[4+2*i] = byte(uint8(mon.x))
		msg[4+2*i+1] = byte(uint8(mon.y))
		msg[4+2*i+4] = byte(uint8(mon.count))
		i++
	}
}

func (s *server) runP(c net.Conn, id int) {
	buf := make([]byte, 10)
	c.Read(buf[:3])
	if bytes.Compare(buf[:3], []byte("NME")) != 0 {
		fmt.Errorf("Invalid first connexion value")
		return
	}

	c.Read(buf[:1])
	t := int(buf[0])
	if t > 10 {
		buf = make([]byte, t)
	}
	c.Read(buf[:t])
	s.name[id] = string(buf[:t])
	s.set(c)
	s.hum(c)
	s.hme(c, id)
	s.state(c, "MAP")
}
