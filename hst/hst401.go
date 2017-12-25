package hst

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/adyzng/go-duka/core"
	"github.com/adyzng/go-duka/misc"
)

var (
	log = misc.NewLogger("HST", 3)
)

// HST401 MT4 history data format .hst with version 401
//
type HST401 struct {
	wg       sync.WaitGroup
	header   *Header
	dest     string
	symbol   string
	spread   uint32
	timefame uint32
	chBars   chan *BarData
	close    chan struct{}
}

// NewHST create a HST convertor
//
func NewHST(timefame, spread uint32, symbol, dest string) *HST401 {
	hst := &HST401{
		header:   NewHeader(timefame, symbol),
		dest:     dest,
		symbol:   symbol,
		spread:   spread,
		timefame: timefame,
		chBars:   make(chan *BarData, 128),
		close:    make(chan struct{}, 1),
	}

	hst.wg.Add(1)
	go hst.worker()

	return hst
}

// worker goroutine which flust data to disk
//
func (h *HST401) worker() error {
	fname := fmt.Sprintf("%s%d.hst", h.symbol, h.timefame)
	fpath := filepath.Join(h.dest, fname)

	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 666)
	if err != nil {
		log.Error("Failed to create file %s, error %v.", fpath, err)
		return err
	}

	defer func() {
		f.Close()
		h.wg.Done()
		log.Trace("Saved : %s.", fpath)
	}()

	// write HST header
	var bs []byte

	if bs, err = h.header.ToBytes(); err != nil {
		log.Error("Pack HST Header (%v) failed: %v.", h.header, err)
		return err
	}
	if _, err = f.Write(bs); err != nil {
		log.Error("Write HST Header (%v) failed: %v.", h.header, err)
		return err
	}

	for bar := range h.chBars {
		if bs, err = bar.ToBytes(); err == nil {
			if _, err = f.Write(bs); err != nil {
				log.Error("Write BarData(%v) failed: %v.", bar, err)
			}
		} else {
			log.Error("Pack BarData(%v) failed: %v.", bar, err)
			continue
		}
	}

	if err != nil {
		log.Warn("HST worker return with %v.", err)
	}
	return err
}

// PackTicks aggregate ticks with timeframe
//
func (h *HST401) PackTicks(barTimestamp uint32, ticks []*core.TickData) error {
	// Transform universal bar list to binary bar data (60 Bytes per bar)
	if len(ticks) == 0 {
		return nil
	}

	bar := &BarData{
		CTM:   uint32(ticks[0].Timestamp / 1000),
		Open:  ticks[0].Bid,
		Low:   ticks[0].Bid,
		High:  ticks[0].Bid,
		Close: ticks[0].Bid,
	}

	for _, tick := range ticks {
		bar.Close = tick.Bid
		bar.Low = math.Min(tick.Bid, bar.Low)
		bar.High = math.Max(tick.Bid, bar.High)
		bar.Volume = bar.Volume + uint64(tick.VolumeAsk+tick.VolumeBid)
	}

	select {
	case h.chBars <- bar:
		break
		//case <-h.close:
		//	break
	}
	return nil
}

// Finish HST file convert
//
func (h *HST401) Finish() error {
	//close(h.close)
	close(h.chBars)
	h.wg.Wait()
	close(h.close)
	return nil
}