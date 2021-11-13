package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/AkinoKaede/naruse/common/bytespool"
	"github.com/AkinoKaede/naruse/vmess"

	"github.com/v2fly/v2ray-core/v4/common/drain"
	"github.com/v2fly/v2ray-core/v4/common/protocol"

	"github.com/database64128/tfo-go"
)

type Dispatcher struct {
	sync.RWMutex

	listener net.Listener

	ListenAddr  string
	Port        int
	TCPFastOpen bool
	Validator   *vmess.Validator
}

func (d *Dispatcher) Listen() error {
	listenConfig := tfo.ListenConfig{
		DisableTFO: !d.TCPFastOpen,
	}
	var err error
	d.listener, err = listenConfig.Listen(context.Background(), "tcp", fmt.Sprintf("%s:%d", d.ListenAddr, d.Port))
	if err != nil {
		return err
	}
	defer d.listener.Close()

	log.Printf("listen on %s:%v\n", d.ListenAddr, d.Port)

	for {
		conn, err := d.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Printf("[error] ReadFrom: %v", err)
			continue
		}
		go func() {
			err := d.handleConn(conn)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}

func (d *Dispatcher) Close() (err error) {
	log.Printf("closed %s:%v\n", d.ListenAddr, d.Port)
	return d.listener.Close()
}

func (d *Dispatcher) UpdateValidator(validator *vmess.Validator) {
	d.Lock()
	defer d.Unlock()
	d.Validator = validator
}

func (d *Dispatcher) handleConn(conn net.Conn) error {
	defer conn.Close()

	d.RLock()
	drainer, err := drain.NewBehaviorSeedLimitedDrainer(int64(d.Validator.GetBehaviorSeed()), 16+38, 3266, 64)
	if err != nil {
		return err
	}
	d.RUnlock()

	data := bytespool.Get(16)
	defer bytespool.Put(data)

	n, err := io.ReadAtLeast(conn, data, protocol.IDBytesLen)
	if err != nil {
		return fmt.Errorf("%s <-x-> %s handleConn ReadAtLeast error: %w", conn.RemoteAddr(), conn.LocalAddr(), err)
	}
	drainer.AcknowledgeReceive(n)

	d.RLock()
	account, err := d.Validator.Get(data[:protocol.IDBytesLen])
	d.RUnlock()
	if err != nil {
		return drain.WithError(drainer, conn, err)
	}

	dialer := tfo.Dialer{
		DisableTFO: !account.Server.TCPFastOpen,
	}

	remoteConn, err := dialer.Dial("tcp", account.Server.Target)
	if err != nil {
		return fmt.Errorf("%s <-> %s <-x-> %s handleConn dial error: %w", conn.RemoteAddr(), conn.LocalAddr(), account.Server.Target, err)
	}

	_, err = remoteConn.Write(data[:n])
	if err != nil {
		return fmt.Errorf("%s <-> %s <-x-> %s handleConn write error: %w", conn.RemoteAddr(), conn.LocalAddr(), account.Server.Target, err)
	}

	log.Printf("%s <-> %s <-> %s", conn.RemoteAddr(), conn.LocalAddr(), account.Server.Target)

	if err := relay(conn.(DuplexConn), remoteConn.(DuplexConn)); err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil // ignore i/o timeout
		}
		return fmt.Errorf("handleConn relay error: %w", err)
	}

	return nil
}

func relay(localConn, remoteConn DuplexConn) error {
	defer remoteConn.Close()
	ch := make(chan error, 1)
	go func() {
		_, err := io.Copy(localConn, remoteConn)
		localConn.CloseWrite()
		ch <- err
	}()
	_, err := io.Copy(remoteConn, localConn)
	remoteConn.CloseWrite()
	innerErr := <-ch
	if err != nil {
		return err
	}
	return innerErr
}
