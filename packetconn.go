package freconn

import (
	"context"
	"net"
)

type PacketConn struct {
	net.PacketConn
	*advanced
}

func WrapPacketConn(pc net.PacketConn, opts ...OptionHandler) *PacketConn {
	newPc := &PacketConn{PacketConn: pc, advanced: &advanced{Flag: 0}}
	for _, h := range opts {
		h(newPc.advanced)
	}
	return newPc
}

func (c *PacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, a, err := c.PacketConn.ReadFrom(b)
	if err != nil {
		return n, a, err
	}
	if haveFlags(c.Flag, FlagRatelimit) && c.RxBucket != nil {
		c.RxBucket.WaitN(context.Background(), n*8)
	}
	if haveFlags(c.Flag, FlagStat) && c.Stat != nil {
		c.Stat.AddRx(uint64(n) * 8)
	}

	return n, a, err
}

func (c *PacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if haveFlags(c.Flag, FlagRatelimit) && c.TxBucket != nil {
		c.TxBucket.WaitN(context.Background(), len(b)*8)
	}
	n, err := c.PacketConn.WriteTo(b, addr)
	if err != nil {
		return n, err
	}
	if haveFlags(c.Flag, FlagStat) && c.Stat != nil {
		c.Stat.AddTx(uint64(n) * 8)
	}

	return n, nil
}
