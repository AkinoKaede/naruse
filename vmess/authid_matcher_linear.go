package vmess

import (
	"hash/crc32"
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

		return v.account, nil
	}
	return nil, ErrNotFound
}
