package vmess

import (
	"hash/crc32"
	"math"
	"time"

	"github.com/v2fly/v2ray-core/v4/common/antireplay"
)

type AuthIDLinearMatcher struct {
	decoders       map[string]*AuthIDDecoderItem
	timestampCheck bool
}

func (a *AuthIDLinearMatcher) AddUser(key [16]byte, account *Account) {
	a.decoders[string(key[:])] = NewAuthIDDecoderItem(key, account)
}

func (a *AuthIDLinearMatcher) RemoveUser(key [16]byte) {
	delete(a.decoders, string(key[:]))
}

func (a *AuthIDLinearMatcher) Match(authID [16]byte) (*Account, error) {
	for _, v := range a.decoders {
		t, z, _, d := v.dec.Decode(authID)
		if z != crc32.ChecksumIEEE(d[:12]) {
			continue
		}

		if t < 0 {
			continue
		}

		if a.timestampCheck && math.Abs(math.Abs(float64(t))-float64(time.Now().Unix())) > 120 {
			continue
		}

		return v.account, nil
	}
	return nil, ErrNotFound
}

type AuthIDLinearMatcherWithReplayFilter struct {
	AuthIDLinearMatcher
	filter *antireplay.ReplayFilter
}

func (a *AuthIDLinearMatcherWithReplayFilter) Match(authID [16]byte) (*Account, error) {
	for _, v := range a.decoders {
		t, z, _, d := v.dec.Decode(authID)
		if z != crc32.ChecksumIEEE(d[:12]) {
			continue
		}

		if t < 0 {
			continue
		}

		if a.timestampCheck && math.Abs(math.Abs(float64(t))-float64(time.Now().Unix())) > 120 {
			continue
		}

		if !a.filter.Check(authID[:]) {
			return nil, ErrReplay
		}

		return v.account, nil
	}
	return nil, ErrNotFound
}

func NewAuthIDLinearMatcher(timestampCheck bool) *AuthIDLinearMatcher {
	return &AuthIDLinearMatcher{make(map[string]*AuthIDDecoderItem), timestampCheck}
}

func NewAuthIDLinearMatcherWithReplayFilter(timestampCheck bool) *AuthIDLinearMatcherWithReplayFilter {
	return &AuthIDLinearMatcherWithReplayFilter{AuthIDLinearMatcher{make(map[string]*AuthIDDecoderItem), timestampCheck}, antireplay.NewReplayFilter(120)}
}
