package ddd

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
	isMIMobile := NewSpecification[Mobile](func(t Mobile) bool {
		return t.Brand == MI
	})
	isVIVOMobile := NewSpecification[Mobile](func(t Mobile) bool {
		return t.Brand == VIVO
	})
	isOPPOMobile := NewSpecification[Mobile](func(t Mobile) bool {
		return t.Brand == OPPO
	})
	isSamSungMobile := NewSpecification[Mobile](func(t Mobile) bool {
		return t.Brand == Samsung
	})

	a := Mobile{Brand: MI}
	assert.True(t, isMIMobile.IsSatisfiedBy(a))
	assert.False(t, isVIVOMobile.IsSatisfiedBy(a))
	assert.False(t, isOPPOMobile.IsSatisfiedBy(a))
	assert.False(t, isSamSungMobile.IsSatisfiedBy(a))

	assert.False(t, NewAndSpecification(isMIMobile, isVIVOMobile).IsSatisfiedBy(a))
	assert.False(t, NewAndSpecification(isOPPOMobile, isSamSungMobile).IsSatisfiedBy(a))

	assert.True(t, NewOrSpecification(isMIMobile, isVIVOMobile).IsSatisfiedBy(a))
	assert.False(t, NewOrSpecification(isOPPOMobile, isSamSungMobile).IsSatisfiedBy(a))

	assert.False(t, NewNotSpecification(isMIMobile).IsSatisfiedBy(a))
	assert.True(t, NewNotSpecification(isVIVOMobile).IsSatisfiedBy(a))
	assert.True(t, NewNotSpecification(isOPPOMobile).IsSatisfiedBy(a))
	assert.True(t, NewNotSpecification(isSamSungMobile).IsSatisfiedBy(a))

}
