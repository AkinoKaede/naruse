package vmess

import (
	"hash/crc32"
	"math"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/antireplay"
)

type AuthIDLinearMatcher struct {
	decoders map[string]*AuthIDDecoderItem
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

		if math.Abs(math.Abs(float64(t))-float64(time.Now().Unix())) > 120 {
			continue
		}

		return v.account, nil
	}
	return nil, ErrNotFound
}

type AuthIDLinearMatcherWithAntiReplay struct {
	AuthIDLinearMatcher
	filter *antireplay.ReplayFilter
}

func (a *AuthIDLinearMatcherWithAntiReplay) Match(authID [16]byte) (*Account, error) {
	for _, v := range a.decoders {
		t, z, _, d := v.dec.Decode(authID)
		if z != crc32.ChecksumIEEE(d[:12]) {
			continue
		}

		if t < 0 {
			continue
		}

		if math.Abs(math.Abs(float64(t))-float64(time.Now().Unix())) > 120 {
			continue
		}

		if !a.filter.Check(authID[:]) {
			return nil, ErrReplay
		}

		return v.account, nil
	}
	return nil, ErrNotFound
}

func NewAuthIDLinearMatcher() AuthIDMatcher {
	return &AuthIDLinearMatcher{make(map[string]*AuthIDDecoderItem)}
}

func NewAuthIDLinearMatcherWithAntiReplay() AuthIDMatcher {
	return &AuthIDLinearMatcherWithAntiReplay{AuthIDLinearMatcher{make(map[string]*AuthIDDecoderItem)}, antireplay.NewReplayFilter(120)}
}
