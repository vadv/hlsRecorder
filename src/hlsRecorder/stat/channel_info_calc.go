package stat

import (
	percentile "github.com/montanaflynn/stats"
)

func (i *ChannelInfo) CalcStat() {

	i.PlayList.Lock()
	i.PlayList.Percentile75, _ = percentile.Percentile(i.PlayList.list, 75)
	i.PlayList.Percentile75 = i.PlayList.Percentile75 * 100
	i.PlayList.Percentile95, _ = percentile.Percentile(i.PlayList.list, 95)
	i.PlayList.Percentile95 = i.PlayList.Percentile95 * 100
	i.PlayList.Percentile99, _ = percentile.Percentile(i.PlayList.list, 99)
	i.PlayList.Percentile99 = i.PlayList.Percentile99 * 100
	i.PlayList.Unlock()

	i.Data.Lock()
	i.Data.Percentile75, _ = percentile.Percentile(i.Data.list, 75)
	i.Data.Percentile75 = i.Data.Percentile75 * 100
	i.Data.Percentile95, _ = percentile.Percentile(i.Data.list, 95)
	i.Data.Percentile95 = i.Data.Percentile95 * 100
	i.Data.Percentile99, _ = percentile.Percentile(i.Data.list, 99)
	i.Data.Percentile99 = i.Data.Percentile99 * 100
	i.Data.Unlock()

}
