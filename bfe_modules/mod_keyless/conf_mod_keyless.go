package mod_keyless

import (
	"fmt"
	"github.com/baidu/bfe/bfe_util"
	"gopkg.in/gcfg.v1"
)

type ConfModKeyless struct {
	Basic struct {
		Server     string
		ServerCa   string
		ServerKey  string
		ServerCert string
		CertsDir   string
	}
}

func (cfg *ConfModKeyless) Check(confRoot string) error {
	if cfg.Basic.Server == "" {
		return fmt.Errorf("invalid keyless server")
	}

	if cfg.Basic.ServerCa == "" {
		return fmt.Errorf("invalid keyless server ca")
	}

	if cfg.Basic.ServerKey == "" {
		return fmt.Errorf("invalid keyless server key")
	}

	if cfg.Basic.ServerCert == "" {
		return fmt.Errorf("invalid keyless server cert")
	}

	if cfg.Basic.CertsDir == "" {
		return fmt.Errorf("invalid keyless cert dir")
	}

	cfg.Basic.ServerCa = bfe_util.ConfPathProc(cfg.Basic.ServerCa, confRoot)
	cfg.Basic.ServerKey = bfe_util.ConfPathProc(cfg.Basic.ServerKey, confRoot)
	cfg.Basic.ServerCert = bfe_util.ConfPathProc(cfg.Basic.ServerCert, confRoot)
	cfg.Basic.CertsDir = bfe_util.ConfPathProc(cfg.Basic.CertsDir, confRoot)
	return nil
}

func ConfLoad(filePaht string, confRoot string) (*ConfModKeyless, error) {
	var cfg ConfModKeyless
	var err error

	err = gcfg.ReadFileInto(&cfg, filePaht)
	if err != nil {
		return nil, err
	}

	// check config
	err = cfg.Check(confRoot)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
