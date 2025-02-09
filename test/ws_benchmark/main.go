package main

import "flag"

var (
	// 起始phone_num
	pn = flag.Int64("pn", 100000, "First phone num")
	// 群成员总数
	gn = flag.Int64("gn", 100, "群成员总数")
	// 在线成员数量
	on = flag.Int64("on", 100, "在线成员数量")
	// 每次发送消息数量（不同在线成员发送一次，总共 500 个在线成员每人发送一次，总共发送 500 次）
	sn = flag.Int64("sn", 100, "每次发送消息数量")
	// 发送消息次数
	tn = flag.Int64("tn", 40, "发送消息次数")
)

func main() {
	flag.Parse()

	mgr := NewManager(*pn, *on, *sn, *gn, *tn)
	//mgr.RunMaxOnlineCount()

	mgr.Run()

	select {}
}
