package prototype

type options struct {
	Copy          func(tgt any, src any) error
	BoolConverter BoolConverter
}

type Option func(o *options)

func Copy(f func(tgt any, src any) error) Option {
	return func(o *options) {
		o.Copy = f
	}
}

func Clone(tgt any, src any, opts ...Option) error {
	o := &options{}
	for _, option := range opts {
		option(o)
	}
	if o.Copy != nil {
		return o.Copy(tgt, src)
	}
	return clone(tgt, src, o)
}

func clone(tgt any, src any, opt *options) (err error) {
	return nil
}
