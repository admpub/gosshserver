package gosshserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

func genkey() (privateKeyBlock *pem.Block, publicKeyBlock *pem.Block, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("不能生成密钥: %w", err)
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("不能解析公钥: %w", err)
	}
	publicKeyBlock = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	return
}

func savekey(keyPath string, privateKeyBlock *pem.Block, publicKeyBlock *pem.Block) error {

	f, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	err = pem.Encode(f, privateKeyBlock)
	f.Close()
	if err != nil {
		return err
	}

	f, err = os.Create(keyPath + `.pub`)
	if err != nil {
		return err
	}
	err = pem.Encode(f, publicKeyBlock)
	f.Close()

	return err
}

func getPrivatePEM(keyPath string) ([]byte, error) {
	b, err := os.ReadFile(keyPath)
	if err == nil {
		return b, err
	}
	if !os.IsNotExist(err) {
		return b, err
	}
	privateKeyBlock, publicKeyBlock, err := genkey()
	if err != nil {
		return b, err
	}
	err = savekey(keyPath, privateKeyBlock, publicKeyBlock)
	if err != nil {
		return b, err
	}
	b = pem.EncodeToMemory(privateKeyBlock)
	return b, err
}

func parsePublicKeyPEM(publicKeyPEM []byte) (ssh.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, errors.New("无效的 RSA 公钥")
	}

	// 解析 RSA 公钥
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		pubKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析公钥失败: %w", err)
		}
	}

	// 转换为 OpenSSH 格式
	sshPubKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("转换为 OpenSSH 格式失败: %w", err)
	}
	return sshPubKey, err
}

// 转换为 OpenSSH 格式
func PublicKeyPEM2AuthorizedKey(publicKeyPEM []byte) ([]byte, error) {
	sshPubKey, err := parsePublicKeyPEM(publicKeyPEM)
	if err != nil {
		return nil, err
	}

	// 输出 OpenSSH 格式的公钥
	return ssh.MarshalAuthorizedKey(sshPubKey), nil
}
