package freconn

import (
	"context"
	"net"

	"golang.org/x/time/rate"
)

const (
	FlagRatelimit = 1 << iota
	FlagStat
)

type Conn struct {
	net.Conn
	*advanced
}

type advanced struct {
	Flag     int
	TxBucket *rate.Limiter
	RxBucket *rate.Limiter
	Stat     *Stat
}

type OptionHandler func(*advanced)

func WithLimit(txBucket, rxBucket *rate.Limiter) OptionHandler {
	return func(adv *advanced) {
		adv.TxBucket = txBucket
		adv.RxBucket = rxBucket
		adv.Flag = adv.Flag | FlagRatelimit
	}
}

func WithStat(stat *Stat) OptionHandler {
	return func(adv *advanced) {
		if stat != nil {
			adv.Stat = stat
			adv.Flag = adv.Flag | FlagRatelimit
		}
	}
}

func WrapConn(c net.Conn, opts ...OptionHandler) *Conn {
	conn := &Conn{Conn: c, advanced: &advanced{Flag: 0}}
	for _, h := range opts {
		h(conn.advanced)
	}
	return conn
}

func haveFlags(flags, want int) bool {
	return flags&want != 0
}

func (c *Conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if err != nil {
		return n, err
	}
	if haveFlags(c.Flag, FlagRatelimit) && c.RxBucket != nil {
		c.RxBucket.WaitN(context.Background(), n*8)
	}
	if haveFlags(c.Flag, FlagStat) && c.Stat != nil {
		c.Stat.AddRx(uint64(n) * 8)
	}
	return n, err
}

func (c *Conn) Write(b []byte) (int, error) {
	if haveFlags(c.Flag, FlagRatelimit) && c.TxBucket != nil {
		c.TxBucket.WaitN(context.Background(), len(b)*8)
	}
	n, err := c.Conn.Write(b)
	if err != nil {
		return n, err
	}
	if haveFlags(c.Flag, FlagStat) && c.Stat != nil {
		c.Stat.AddTx(uint64(n) * 8)
	}
	return n, nil
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}
