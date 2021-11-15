package vmess

import (
	"hash/crc32"
	"math"
	"time"
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

func NewAuthIDLinearMatcher(timestampCheck bool) *AuthIDLinearMatcher {
	return &AuthIDLinearMatcher{make(map[string]*AuthIDDecoderItem), timestampCheck}
}
