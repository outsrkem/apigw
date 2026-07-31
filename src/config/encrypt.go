package config

import (
	"apigw/src/cfgtypes"
	"apigw/src/pkg/crypto"
	"fmt"
	"os"
)

// encryption encrypts plaintext string and print result, then exit program
func encryption(plain string) {
	fmt.Println(crypto.Encryption(plain))
	os.Exit(0)
}

// decryptionRedisPwd decrypt encrypted redis password in config structure
func decryptionRedisPwd(c *cfgtypes.Config) {
	if c.Apigw.Redis.Password != "" {
		plain, err := crypto.Decryption(c.Apigw.Redis.Password)
		if err != nil {
			fmt.Printf("Redis password decryption failed, ciphertext: %s", c.Apigw.Redis.Password)
			os.Exit(100)
		}
		c.Apigw.Redis.Password = plain
	}
}

// decryCfgCipher decrypt ciphertext string pointer from yaml configuration
func decryCfgCipher(cipherText *string) {
	if cipherText == nil || *cipherText == "" {
		return
	}
	plain, err := crypto.Decryption(*cipherText)
	if err != nil {
		fmt.Printf("Password decryption failed, ciphertext: %s", *cipherText)
		os.Exit(100)
	}
	*cipherText = plain
}
