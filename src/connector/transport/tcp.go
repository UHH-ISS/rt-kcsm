package transport

import (
	"net"
	"rtkcsm/component/behaviour"
	"rtkcsm/connector/reader"
)

type TcpTransport struct {
	ListenAddress string
}

func (transport *TcpTransport) Start(rtkcsm behaviour.RTKCSM, reader reader.AlertReader) error {
	listener, err := net.Listen("tcp", transport.ListenAddress)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			return err
		}

		go transport.hanldeConnection(connection, rtkcsm, reader)
	}
}

func (transport *TcpTransport) hanldeConnection(connection net.Conn, rtkcsm behaviour.RTKCSM, reader reader.AlertReader) error {
	defer connection.Close()
	return reader.ChannelAlerts(rtkcsm, connection)
}
