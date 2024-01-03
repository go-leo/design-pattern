package prototype

// ClonerFrom 自定义克隆方法，从源克隆到自己
type ClonerFrom interface {
	// CloneFrom 如果返回一个错误，则终止clone, 如果bool返回true，则代表当前克隆结束，如果返回false，则代表没有克隆，需要继续克隆。
	CloneFrom(src any) (bool, error)
}

// ClonerTo 自定义克隆方法，将自己克隆到目标
type ClonerTo interface {
	// CloneTo 如果返回一个错误，则终止clone, 如果bool返回true，则代表当前克隆结束，如果返回false，则代表没有克隆结束，需要继续克隆。
	CloneTo(tgt any) (bool, error)
}
