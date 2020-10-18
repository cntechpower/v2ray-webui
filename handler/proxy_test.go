package handler

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/go-playground/validator/v10"
)

func TestDomainNameValidator(t *testing.T) {
	v := validator.New()
	assert.Equal(t, v.Var("cntechpower.com", "required,fqdn"), nil)
	assert.NotEqual(t, v.Var("  ", "required,fqdn"), nil)
	assert.NotEqual(t, v.Var("", "required,fqdn"), nil)
	assert.NotEqual(t, v.Var("   cntechpower.com", "required,fqdn"), nil)
}
