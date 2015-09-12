/*
nexus包为cellnet提供了跨进程,机器的访问支持

每个独立操作系统进程就是一个region, 通过配置文件设定region间的互联方法


*/
package nexus

import (
	"github.com/davyxu/cellnet/dispatcher"
	"github.com/davyxu/cellnet/proto/coredef"
)

var disp = dispatcher.NewDataDispatcher()

func init() {

	dispatcher.AddMapper(coredef.RegionLinkACK{})

	register(disp)

	listenNexus()

	joinAddr := config.Join

	if joinAddr != "" {

		joinNexus(joinAddr)
	}
}
