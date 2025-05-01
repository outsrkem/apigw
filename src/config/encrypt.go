package config

import (
	"apigw/src/cfgtypts"
	"apigw/src/pkg/crypto"
	"apigw/src/slog"
	"fmt"
	"os"
)

func encryption(plain string) {
	fmt.Println(crypto.Encryption(plain))
	os.Exit(0)
}

func decryptionRedisPwd(c *cfgtypts.Config) {
	klog := slog.FromContext(nil)
	if c.Apigw.Redis.Password != "" {
		if plain, err := crypto.Decryption(c.Apigw.Redis.Password); err != nil {
			klog.Errorf("Decryption of database password failed. uias.yaml:uias.database.passwd %s", c.Apigw.Redis.Password)
			os.Exit(100)
		} else {
			c.Apigw.Redis.Password = plain
		}
	}
}
