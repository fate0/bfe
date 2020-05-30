package mod_keyless

// bfe_tls require
// https://github.com/golang/go/commit/28f33b4a7071870e5ee8b3f87170bbdf9c08981e
// https://github.com/golang/go/commit/7b850ec6917acada87482bbdea76abb57aa5f9cd

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/baidu/bfe/bfe_module"
	"github.com/baidu/bfe/bfe_tls"
	"github.com/baidu/go-lib/web-monitor/web_monitor"
	"github.com/cloudflare/gokeyless/client"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	ModKeyless = "mod_keyless"
)

type ModuleKeyless struct {
	name     string
	client   *client.Client
	conf     *ConfModKeyless
	certsMap map[string]*bfe_tls.Certificate
}

func NewModuleKeyless() *ModuleKeyless {
	m := new(ModuleKeyless)
	m.name = ModKeyless
	return m
}

func (m *ModuleKeyless) Name() string {
	return m.name
}

func (m *ModuleKeyless) Init(cbs *bfe_module.BfeCallbacks, whs *web_monitor.WebHandlers,
	cr string) error {
	var conf *ConfModKeyless
	var err error

	m.certsMap = make(map[string]*bfe_tls.Certificate)

	// load config
	confPath := bfe_module.ModConfPath(cr, m.name)
	if conf, err = ConfLoad(confPath, cr); err != nil {
		return fmt.Errorf("%s: conf load err %s", m.name, err.Error())
	}
	m.conf = conf

	m.client, err = client.NewClientFromFile(m.conf.Basic.ServerCert, m.conf.Basic.ServerKey, m.conf.Basic.ServerCa)
	if err != nil {
		return fmt.Errorf("create keyless client failed %s", err.Error())
	}
	filepath.Walk(m.conf.Basic.CertsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		isCert := strings.HasSuffix(info.Name(), "crt")

		if !info.IsDir() && isCert {
			c, err := m.loadTLSCertificate(m.conf.Basic.Server, path)
			if err != nil {
				return err
			}
			m.certsMap[strings.TrimSuffix(info.Name(), ".crt")] = &c
		}

		return nil
	})

	bfe_tls.SetTlsMultiCertificate(*m)
	return nil
}

func (m *ModuleKeyless) loadTLSCertificate(server, certFile string) (cert bfe_tls.Certificate, err error) {
	fail := func(err error) (bfe_tls.Certificate, error) { return bfe_tls.Certificate{}, err }
	var certPEMBlock []byte
	var certDERBlock *pem.Block

	if certPEMBlock, err = ioutil.ReadFile(certFile); err != nil {
		return fail(err)
	}

	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}

		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}

	if len(cert.Certificate) == 0 {
		return fail(fmt.Errorf("crypto/tls: failed to parse certificate PEM data"))
	}

	if cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
		return fail(err)
	}

	cert.PrivateKey, err = m.client.NewRemoteSignerByCert(server, cert.Leaf)
	if err != nil {
		return fail(err)
	}

	return cert, nil
}

func (m ModuleKeyless) Get(c *bfe_tls.Conn) *bfe_tls.Certificate {
	if cert, ok := m.certsMap[c.GetServerName()]; ok {
		return cert
	}

	return nil
}
