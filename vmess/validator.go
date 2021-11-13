package vmess

import (
	"sync"

	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/uuid"
)

type Validator struct {
	sync.RWMutex

	AuthIDMatcher AuthIDMatcher
}

func (v *Validator) Add(account *Account) {
	v.Lock()
	defer v.Unlock()

	var cmdKeyFL [16]byte
	copy(cmdKeyFL[:], account.ID.CmdKey())
	v.AuthIDMatcher.AddUser(cmdKeyFL, account)
}

func (v *Validator) Remove(uuid uuid.UUID) {
	v.Lock()
	defer v.Unlock()

	id := protocol.NewID(uuid)
	var cmdKeyFL [16]byte
	copy(cmdKeyFL[:], id.CmdKey())
	v.AuthIDMatcher.RemoveUser(cmdKeyFL)
}

func (v *Validator) Get(userHash []byte) (*Account, error) {
	v.RLock()
	defer v.RUnlock()

	var userHashFL [16]byte
	copy(userHashFL[:], userHash)

	return v.AuthIDMatcher.Match(userHashFL)
}
