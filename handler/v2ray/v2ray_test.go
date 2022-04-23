package v2ray

import (
	"io/ioutil"
	"testing"

	"github.com/cntechpower/utils/log"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionParse(t *testing.T) {
	log.InitLogger("")
	s, err := New("/tmp/v2ray.txt")
	if !assert.Equal(t, nil, err) {
		t.FailNow()
	}
	bs, err := ioutil.ReadFile("/tmp/v2ray.txt")
	if !assert.Equal(t, nil, err) {
		t.FailNow()
	}
	res, err := s.decodeSubscription(1, "dounai", bs)
	assert.Equal(t, nil, err)
	for _, r := range res {
		t.Logf("%+v\n", r)
	}
}
