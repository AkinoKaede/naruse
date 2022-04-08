package vmess

import (
	"github.com/v2fly/v2ray-core/v5/proxy/vmess/aead"
)

type AuthIDDecoderItem struct {
	dec     *aead.AuthIDDecoder
	account *Account
}

func NewAuthIDDecoderItem(key [16]byte, account *Account) *AuthIDDecoderItem {
	return &AuthIDDecoderItem{
		dec:     aead.NewAuthIDDecoder(key[:]),
		account: account,
	}
}
