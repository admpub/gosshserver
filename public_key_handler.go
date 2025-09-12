package gosshserver

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/admpub/log"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func isDebug() bool {
	return log.IsEnabled(log.LevelDebug)
}

func certificateVerify(cfg Config, ctx ssh.Context, key ssh.PublicKey, cert *gossh.Certificate) bool {
	if len(cfg.TrustedUserCAKeys) == 0 {
		log.Warn("Certificate Rejected: No trusted certificate authorities for this server")
		log.Warnf("Failed authentication attempt from %s", ctx.RemoteAddr())
		return false
	}

	if cert.CertType != gossh.UserCert {
		log.Warn("Certificate Rejected: Not a user certificate")
		log.Warnf("Failed authentication attempt from %s", ctx.RemoteAddr())
		return false
	}

	for _, principal := range cert.ValidPrincipals {
		c := &gossh.CertChecker{
			IsUserAuthority: func(auth gossh.PublicKey) bool {
				marshaled := auth.Marshal()
				for _, k := range cfg.TrustedUserCAKeysParsed {
					if bytes.Equal(marshaled, k.Marshal()) {
						return true
					}
				}

				return false
			},
		}

		// check the CA of the cert
		if !c.IsUserAuthority(cert.SignatureKey) {
			if isDebug() {
				log.Debugf("Principal Rejected: %s Untrusted Authority Signature Fingerprint %s for Principal: %s", ctx.RemoteAddr(), gossh.FingerprintSHA256(cert.SignatureKey), principal)
			}
			continue
		}

		// validate the cert for this principal
		if err := c.CheckCert(principal, cert); err != nil {
			// User is presenting an invalid certificate - STOP any further processing
			log.Errorf("Invalid Certificate KeyID %s with Signature Fingerprint %s presented for Principal: %s from %s", cert.KeyId, gossh.FingerprintSHA256(cert.SignatureKey), principal, ctx.RemoteAddr())
			log.Warnf("Failed authentication attempt from %s", ctx.RemoteAddr())

			return false
		}

		if isDebug() { // <- FingerprintSHA256 is kinda expensive so only calculate it if necessary
			log.Debugf("Successfully authenticated: %s Certificate Fingerprint: %s Principal: %s", ctx.RemoteAddr(), gossh.FingerprintSHA256(key), principal)
		}
		return true
	}

	log.Warnf("From %s Fingerprint: %s is a certificate, but no valid principals found", ctx.RemoteAddr(), gossh.FingerprintSHA256(key))
	log.Warnf("Failed authentication attempt from %s", ctx.RemoteAddr())
	return false
}

func publicKeyHandler(cfg Config) func(ctx ssh.Context, key ssh.PublicKey) bool {
	authorizedKeysFile := filepath.Join(filepath.Dir(cfg.KeyPath), `authorized_keys`)
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		if ctx.User() != cfg.User {
			log.Warnf("Invalid SSH username %s - must use %s for all git operations via ssh", ctx.User(), cfg.User)
			log.Warnf("Failed authentication attempt from %s", ctx.RemoteAddr())
			return false
		}
		// check if we have a certificate
		if cert, ok := key.(*gossh.Certificate); ok {
			if isDebug() { // <- FingerprintSHA256 is kinda expensive so only calculate it if necessary
				log.Debugf("Handle Certificate: %s Fingerprint: %s is a certificate", ctx.RemoteAddr(), gossh.FingerprintSHA256(key))
			}
			if certificateVerify(cfg, ctx, key, cert) {
				return true
			}
		}

		var allowed ssh.PublicKey
		data, err := os.ReadFile(authorizedKeysFile)
		if err == nil {
			allowed, _, _, _, err = ssh.ParseAuthorizedKey(data)
			if err != nil {
				log.Error(err)
			} else if ssh.KeysEqual(key, allowed) {
				return true
			}
		}
		data, err = os.ReadFile(cfg.KeyPath + `.pub`)
		if err != nil {
			log.Debug(err)
			return false
		}
		allowed, err = parsePublicKeyPEM(data)
		if err != nil {
			log.Error(err)
			return false
		}
		if ssh.KeysEqual(key, allowed) {
			return true
		}
		if isDebug() { // <- FingerprintSHA256 is kinda expensive so only calculate it if necessary
			log.Debugf("Handle Public Key: %s Fingerprint: %s is not a certificate", ctx.RemoteAddr(), gossh.FingerprintSHA256(key))
		}
		return false
	}
}
