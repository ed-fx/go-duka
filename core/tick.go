package core

import (
	"fmt"
	"time"
)

// TickData for dukascopy
//
type TickData struct {
	Symbol    string  // 货币对
	Timestamp int64   // 时间戳(ms)
	Ask       float64 // 卖价
	Bid       float64 // 买价
	VolumeAsk float64 // 单位：MIO(百万)
	VolumeBid float64 // 单位：MIO(百万)
}

// BarData means tick data within one Bar
//
type BarData struct {
	TickTimestamp uint32  // second
	BarTimestamp  uint32  // second
	Open          float64 // OLHCV
	Low           float64 //
	High          float64 //
	Close         float64 //
	Volume        uint64  //
}

// ToString used to format into csv row
//
func (t *TickData) ToString() []string {
	tm := time.Unix(t.Timestamp/1000, (t.Timestamp%1000)*int64(time.Millisecond))
	return []string{
		tm.Format("2006-01-02 15:04:05.000"),
		fmt.Sprintf("%.5f", t.Ask),
		fmt.Sprintf("%.5f", t.Bid),
		fmt.Sprintf("%.2f", t.VolumeAsk),
		fmt.Sprintf("%.2f", t.VolumeBid),
	}
}
