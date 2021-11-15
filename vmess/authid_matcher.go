package vmess

type AuthIDMatcher interface {
	Match([16]byte) (*Account, error)
	AddUser([16]byte, *Account)
	RemoveUser([16]byte)
}

type NewAuthIDMatcherFunc func() AuthIDMatcher

// map[AntiReplay]NewAuthIDMatcherFunc
var NewAuthIDMatchers = map[bool]NewAuthIDMatcherFunc{
	false: NewAuthIDLinearMatcher,
	true:  NewAuthIDLinearMatcherWithAntiReplay,
}
