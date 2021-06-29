package freconn

import (
	"context"
	"sync/atomic"
	"time"
)

type Stat struct {
	Rx     uint64
	Tx     uint64
	bw     bandwidth
	ctx    context.Context
	cancel context.CancelFunc
}

type bandwidth struct {
	lastRx      uint64
	lastTx      uint64
	bandwidthRx uint64
	bandwidthTx uint64
}

func (s *bandwidth) reset() {
	atomic.StoreUint64(&s.lastRx, 0)
	atomic.StoreUint64(&s.lastTx, 0)
	atomic.StoreUint64(&s.bandwidthRx, 0)
	atomic.StoreUint64(&s.bandwidthTx, 0)
}

func NewStat() *Stat {
	return NewStatWithInterval(context.Background(), time.Second)
}

func NewStatWithInterval(ctx context.Context, interval time.Duration) *Stat {
	s := &Stat{
		Rx: 0,
		Tx: 0,
		bw: bandwidth{
			lastRx:      0,
			lastTx:      0,
			bandwidthRx: 0,
			bandwidthTx: 0,
		},
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	go s.runBandwidthInterval(s.ctx, interval)

	return s
}

func (s *Stat) runBandwidthInterval(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	select {
	case <-ctx.Done():
		return
	case <-ticker.C:
		rx := atomic.LoadUint64(&s.Rx)
		tx := atomic.LoadUint64(&s.Tx)
		atomic.StoreUint64(&s.bw.bandwidthRx, rx-atomic.LoadUint64(&s.bw.lastRx))
		atomic.StoreUint64(&s.bw.bandwidthTx, tx-atomic.LoadUint64(&s.bw.lastTx))
		atomic.StoreUint64(&s.bw.lastRx, rx)
		atomic.StoreUint64(&s.bw.lastTx, tx)
		// log.Printf("[1s]RX: %dbps", s.in1.bandwidthRx)
		// log.Printf("[1s]TX: %dbps", s.in1.bandwidthTx)
	}
}

func (s *Stat) GetRx() uint64 {
	return atomic.LoadUint64(&s.Rx)
}

func (s *Stat) GetTx() uint64 {
	return atomic.LoadUint64(&s.Tx)
}

func (s *Stat) AddRx(len uint64) {
	atomic.AddUint64(&s.Rx, len)
}

func (s *Stat) AddTx(len uint64) {
	atomic.AddUint64(&s.Tx, len)
}

func (s *Stat) Bandwidth() (r, t uint64) {
	r = atomic.LoadUint64(&s.bw.bandwidthRx)
	t = atomic.LoadUint64(&s.bw.bandwidthTx)
	return
}

func (s *Stat) Reset() {
	atomic.StoreUint64(&s.Rx, 0)
	atomic.StoreUint64(&s.Tx, 0)
	s.bw.reset()
}

func (s *Stat) Close() {
	s.cancel()
}
