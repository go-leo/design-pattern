package prototype

import (
	"github.com/go-leo/prototype"
)

// Cloner is a clone hook.
type Cloner = prototype.Cloner

type Option = prototype.Option

var TagKey = prototype.TagKey

var DisableDeepClone = prototype.DisableDeepClone

var EqualFold = prototype.EqualFold

var IntToTime = prototype.IntToTime

var StringToTime = prototype.StringToTime

var TimeToInt = prototype.TimeToInt

var TimeToString = prototype.TimeToString

var GetterPrefix = prototype.GetterPrefix

var SetterPrefix = prototype.SetterPrefix

var Context = prototype.Context

var Cloners = prototype.Cloners

var InterruptOnError = prototype.InterruptOnError

// Clone 将值从 src 克隆到 tgt
var Clone = prototype.Clone
