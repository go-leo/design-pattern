package specification

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Mobile struct {
	Brand string
}

const (
	MI      = "xiaomi"
	VIVO    = "vivo"
	OPPO    = "oppo"
	Samsung = "samsung"
)

func TestSpecification(t *testing.T) {
	isMIMobile := New[Mobile](func(ctx context.Context, t Mobile) bool {
		return t.Brand == MI
	})
	isVIVOMobile := New[Mobile](func(ctx context.Context, t Mobile) bool {
		return t.Brand == VIVO
	})
	isOPPOMobile := New[Mobile](func(ctx context.Context, t Mobile) bool {
		return t.Brand == OPPO
	})
	isSamSungMobile := New[Mobile](func(ctx context.Context, t Mobile) bool {
		return t.Brand == Samsung
	})

	a := Mobile{Brand: MI}
	assert.True(t, isMIMobile.IsSatisfiedBy(context.Background(), a))
	assert.False(t, isVIVOMobile.IsSatisfiedBy(context.Background(), a))
	assert.False(t, isOPPOMobile.IsSatisfiedBy(context.Background(), a))
	assert.False(t, isSamSungMobile.IsSatisfiedBy(context.Background(), a))

	assert.False(t, And(isMIMobile, isVIVOMobile).IsSatisfiedBy(context.Background(), a))
	assert.False(t, And(isOPPOMobile, isSamSungMobile).IsSatisfiedBy(context.Background(), a))

	assert.True(t, Or(isMIMobile, isVIVOMobile).IsSatisfiedBy(context.Background(), a))
	assert.False(t, Or(isOPPOMobile, isSamSungMobile).IsSatisfiedBy(context.Background(), a))

	assert.False(t, Not(isMIMobile).IsSatisfiedBy(context.Background(), a))
	assert.True(t, Not(isVIVOMobile).IsSatisfiedBy(context.Background(), a))
	assert.True(t, Not(isOPPOMobile).IsSatisfiedBy(context.Background(), a))
	assert.True(t, Not(isSamSungMobile).IsSatisfiedBy(context.Background(), a))

}
