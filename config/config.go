package config

import (
	"encoding/json"
	"os"

	"github.com/AkinoKaede/naruse/dispatcher"
	"github.com/AkinoKaede/naruse/vmess"

	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/uuid"
)

type Config struct {
	Groups []Group `json:"groups"`
}

type Group struct {
	Listen      string   `json:"listen"`
	Port        int      `json:"port"`
	TCPFastOpen bool     `json:"tcpFastOpen"`
	AntiReplay  bool     `json:"antiReplay"`
	Servers     []Server `json:"servers"`
}

type Server struct {
	Target      string   `json:"target"`
	ID          []string `json:"id"`
	TCPFastOpen bool     `json:"tcpFastOpen"`
}

func BuildConfig(path string) (*Config, error) {
	config := new(Config)

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Build() ([]*dispatcher.Dispatcher, error) {
	dispatchers := make([]*dispatcher.Dispatcher, len(c.Groups))

	for i, g := range c.Groups {
		dispatcher, err := g.Build()
		if err != nil {
			return nil, err
		}

		dispatchers[i] = dispatcher
	}

	return dispatchers, nil
}

func (g *Group) Build() (*dispatcher.Dispatcher, error) {
	accounts, err := g.AsAccounts()
	if err != nil {
		return nil, err
	}

	validator := &vmess.Validator{
		AuthIDMatcher: vmess.NewAuthIDMatchers[g.AntiReplay](),
	}

	for _, account := range accounts {
		validator.Add(account)
	}

	return &dispatcher.Dispatcher{
		ListenAddr:  g.Listen,
		Port:        g.Port,
		TCPFastOpen: g.TCPFastOpen,
		Validator:   validator,
	}, nil
}

func (g *Group) AsAccounts() ([]*vmess.Account, error) {
	accounts := make([]*vmess.Account, 0)

	for _, s := range g.Servers {
		serverAccounts, err := s.AsAccounts()
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, serverAccounts...)
	}

	return accounts, nil
}

func (s *Server) AsAccounts() ([]*vmess.Account, error) {
	accounts := make([]*vmess.Account, len(s.ID))

	for i, id := range s.ID {
		uuid, err := uuid.ParseString(id)
		if err != nil {
			return nil, err
		}

		vID := protocol.NewID(uuid)

		account := &vmess.Account{
			ID: vID,
			Server: &vmess.Server{
				Target:      s.Target,
				TCPFastOpen: s.TCPFastOpen,
			},
		}

		accounts[i] = account
	}

	return accounts, nil
}
