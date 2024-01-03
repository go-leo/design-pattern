package prototype

// ClonerFrom 自定义克隆方法，从源克隆到自己
type ClonerFrom interface {
	CloneFrom(src any) error
}

// ClonerTo 自定义克隆方法，将自己克隆到目标
type ClonerTo interface {
	CloneTo(tgt any) error
}
