package example

import "context"

var _ SourceToTarget = (*SourceToTargetCloner)(nil)

type SourceToTargetCloner struct {
}

func (cloner *SourceToTargetCloner) Clone(ctx context.Context, source *Source) (*Target, error) {
	customCloner, ok := any(cloner).(interface {
		CustomClone(ctx context.Context, source *Source) (*Target, error)
	})
	if ok {
		return customCloner.CustomClone(ctx, source)
	}
	tgt := new(Target)
	beforeCloner, ok := any(cloner).(interface {
		BeforeClone(ctx context.Context, source *Source, tgt *Target) error
	})
	if ok {
		err := beforeCloner.BeforeClone(ctx, source, tgt)
		if err != nil {
			return nil, err
		}
	}

	afterCloner, ok := any(cloner).(interface {
		AfterClone(ctx context.Context, source *Source, tgt *Target) error
	})
	if ok {
		err := afterCloner.AfterClone(ctx, source, tgt)
		if err != nil {
			return nil, err
		}
	}
	return tgt, nil
}
