package vmess

import (
	"crypto/hmac"
	"crypto/sha256"
	"hash/crc64"
	"sync"

	"github.com/v2fly/v2ray-core/v4/common/dice"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/uuid"
)

type Validator struct {
	sync.RWMutex

	AuthIDMatcher AuthIDMatcher

	behaviorSeed  uint64
	behaviorFused bool
}

func (v *Validator) Add(account *Account) {
	v.Lock()
	defer v.Unlock()

	if !v.behaviorFused {
		hashkdf := hmac.New(sha256.New, []byte("VMESSBSKDF"))
		hashkdf.Write(account.ID.Bytes())
		v.behaviorSeed = crc64.Update(v.behaviorSeed, crc64.MakeTable(crc64.ECMA), hashkdf.Sum(nil))
	}

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

func (v *Validator) GetBehaviorSeed() uint64 {
	v.Lock()
	defer v.Unlock()

	v.behaviorFused = true
	if v.behaviorSeed == 0 {
		v.behaviorSeed = dice.RollUint64()
	}
	return v.behaviorSeed
}
