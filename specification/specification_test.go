package specification

import (
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
	isMIMobile := New[Mobile](func(t Mobile) bool {
		return t.Brand == MI
	})
	isVIVOMobile := New[Mobile](func(t Mobile) bool {
		return t.Brand == VIVO
	})
	isOPPOMobile := New[Mobile](func(t Mobile) bool {
		return t.Brand == OPPO
	})
	isSamSungMobile := New[Mobile](func(t Mobile) bool {
		return t.Brand == Samsung
	})

	a := Mobile{Brand: MI}
	assert.True(t, isMIMobile.IsSatisfiedBy(a))
	assert.False(t, isVIVOMobile.IsSatisfiedBy(a))
	assert.False(t, isOPPOMobile.IsSatisfiedBy(a))
	assert.False(t, isSamSungMobile.IsSatisfiedBy(a))

	assert.False(t, And(isMIMobile, isVIVOMobile).IsSatisfiedBy(a))
	assert.False(t, And(isOPPOMobile, isSamSungMobile).IsSatisfiedBy(a))

	assert.True(t, Or(isMIMobile, isVIVOMobile).IsSatisfiedBy(a))
	assert.False(t, Or(isOPPOMobile, isSamSungMobile).IsSatisfiedBy(a))

	assert.False(t, Not(isMIMobile).IsSatisfiedBy(a))
	assert.True(t, Not(isVIVOMobile).IsSatisfiedBy(a))
	assert.True(t, Not(isOPPOMobile).IsSatisfiedBy(a))
	assert.True(t, Not(isSamSungMobile).IsSatisfiedBy(a))

}
