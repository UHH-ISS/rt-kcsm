package transport

import (
	"net"
	"os"
	"os/signal"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"rtkcsm/connector/reader"
	"syscall"
)

type UnixSocketTransport[T structure.Stage, K structure.Stage] struct {
	FilePath string
}

func (transport *UnixSocketTransport[T, K]) Start(rtkcsm behaviour.RTKCSM[T, K], reader reader.AlertReader[T, K]) error {
	listener, err := net.Listen("unix", transport.FilePath)
	if err != nil {
		return err
	}
	defer listener.Close()

	signalChannel := make(chan os.Signal, 1)

	go func() {
		<-signalChannel
		os.Remove(transport.FilePath)
		os.Exit(2)
	}()

	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT)

	for {
		connection, err := listener.Accept()
		if err != nil {
			return err
		}

		go transport.handleConnection(connection, rtkcsm, reader)
	}
}

func (transport *UnixSocketTransport[T, K]) handleConnection(connection net.Conn, rtkcsm behaviour.RTKCSM[T, K], reader reader.AlertReader[T, K]) error {
	defer connection.Close()
	return reader.ChannelAlerts(rtkcsm, connection)
}
