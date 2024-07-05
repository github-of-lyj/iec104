package iec104

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/sirupsen/logrus"
)

func NewServer(address string, tc *tls.Config, lg *logrus.Logger) *Server {
	return &Server{
		address: address,
		tc:      tc,
		lg:      lg,
	}
}

// Server in IEC 104 is also called as slave or controlled station.
type Server struct {
	address  string
	tc       *tls.Config
	listener net.Listener

	lg *logrus.Logger
}

func (s *Server) Serve(handler ClientHandler) error {
	if err := s.listen(); err != nil {
		return err
	}
	defer s.listener.Close()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.lg.Errorf("accept conn with %s", conn.RemoteAddr())
			continue
		}

		go s.serve(&Conn{
			conn,
		}, handler)
	}
}
func (s *Server) listen() error {
	if s.tc != nil {
		listener, err := tls.Listen("tcp", s.address, s.tc)
		if err != nil {
			return err
		}
		s.lg.Debugf("IEC104 server serve at %s with security: %+v", s.address, s.tc)
		s.listener = listener
	} else {
		listener, err := net.Listen("tcp", s.address)
		if err != nil {
			return err
		}
		s.lg.Debugf("IEC104 server serve at %s no security", s.address)
		s.listener = listener
	}
	return nil
}
func (s *Server) serve(conn *Conn, handler ClientHandler) {
	s.lg.Debugf("serve connection from %s", conn.RemoteAddr())
	// TODO
	option, _ := NewClientOption(s.address, handler)
	client := NewClient(option)
	ctx, cancel := context.WithCancel(context.Background())
	client.cancel = cancel
	client.conn = conn
	//用于发送数据
	go client.writingToSocket(ctx)
	//用于接收数据
	go client.readingFromSocket(ctx)

	// var readData = []byte{}
	// for {
	// 	conn.Read(readData)
	// 	s.lg.Printf("读取到的数据：" + string(readData))
	// 	readData = nil
	// 	time.Sleep(2 * time.Second)
	// }

}

type Conn struct {
	net.Conn
}
