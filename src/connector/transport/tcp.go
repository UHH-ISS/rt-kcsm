package transport

import (
	"net"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"rtkcsm/connector/reader"
)

type TcpTransport[T structure.Stage, K structure.Stage] struct {
	ListenAddress string
}

func (transport *TcpTransport[T, K]) Start(rtkcsm behaviour.RTKCSM[T, K], reader reader.AlertReader[T, K]) error {
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

		go transport.handleConnection(connection, rtkcsm, reader)
	}
}

func (transport *TcpTransport[T, K]) handleConnection(connection net.Conn, rtkcsm behaviour.RTKCSM[T, K], reader reader.AlertReader[T, K]) error {
	defer connection.Close()
	return reader.ChannelAlerts(rtkcsm, connection)
}
