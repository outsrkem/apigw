package proxy

import (
	"apigw/src/models"
	"apigw/src/pkg/utils"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strings"
)

func BuildtlsCfg(lcCaID string) (*tls.Config, error) {
	tlsCfg := &tls.Config{}

	dao, err := models.NewDBModel("SYSTEM")
	if err != nil {
		return nil, err
	}
	lcCa, err := dao.GetLcCaByID(lcCaID)
	if err != nil {
		return nil, err
	}

	b64Str := strings.TrimSpace(lcCa.Cert)
	if b64Str == "" {
		tlsCfg.InsecureSkipVerify = true
		return tlsCfg, nil
	}

	cert, err := utils.Base64ToCert(b64Str)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(cert) {
		return nil, errors.New("CA certificate PEM format is invalid")
	}

	tlsCfg.RootCAs = pool
	tlsCfg.InsecureSkipVerify = false
	return tlsCfg, nil
}
