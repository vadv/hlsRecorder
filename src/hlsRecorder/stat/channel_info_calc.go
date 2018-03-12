package stat

import (
	percentile "github.com/montanaflynn/stats"
)

func (i *ChannelInfo) CalcStat() {

	i.PlayList.Lock()
	i.PlayList.Percentile75ms, _ = percentile.Percentile(i.PlayList.list, 75)
	i.PlayList.Percentile75ms = i.PlayList.Percentile75ms * 100
	i.PlayList.Percentile95ms, _ = percentile.Percentile(i.PlayList.list, 95)
	i.PlayList.Percentile95ms = i.PlayList.Percentile95ms * 100
	i.PlayList.Percentile99ms, _ = percentile.Percentile(i.PlayList.list, 99)
	i.PlayList.Percentile99ms = i.PlayList.Percentile99ms * 100
	i.PlayList.Unlock()

	i.Data.Lock()
	i.Data.Percentile75ms, _ = percentile.Percentile(i.Data.list, 75)
	i.Data.Percentile75ms = i.Data.Percentile75ms * 100
	i.Data.Percentile95ms, _ = percentile.Percentile(i.Data.list, 95)
	i.Data.Percentile95ms = i.Data.Percentile95ms * 100
	i.Data.Percentile99ms, _ = percentile.Percentile(i.Data.list, 99)
	i.Data.Percentile99ms = i.Data.Percentile99ms * 100
	i.Data.Unlock()

}
