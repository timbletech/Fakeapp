package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// GenerateChallenge returns a 32-byte crypto-secure random challenge, base64 encoded.
func GenerateChallenge() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// VerifySignature verifies an ECDSA signature over the base64-encoded challenge.
// publicKeyBase64 is expected to be a PEM-encoded ECDSA public key in base64, or just PEM.
// The signature is expected to be ASN.1 encoded (r, s), base64'd.
func VerifySignature(publicKeyBase64 string, challengeBase64 string, signatureBase64 string) (bool, error) {
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		// fallback just in case it's pure PEM without base64 wrapper
		pubBytes = []byte(publicKeyBase64)
	}

	block, _ := pem.Decode(pubBytes)
	if block == nil {
		return false, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse ECDSA public key: %v", err)
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return false, errors.New("not an ECDSA public key")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature base64: %v", err)
	}

	hash := sha256.Sum256([]byte(challengeBase64))

	valid := ecdsa.VerifyASN1(ecdsaPub, hash[:], sigBytes)
	return valid, nil
}

// GenerateKeyPair generates a new ECDSA key pair. 
// Used primarily for DEV mode testing.
// Returns private key block and public key block (PEM Base64).
func GenerateKeyPair() (privB64 string, pubB64 string, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", "", err
	}
	privPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return base64.StdEncoding.EncodeToString(privPem), base64.StdEncoding.EncodeToString(pubPem), nil
}

// SignChallenge signs the challenge for Devsigner mock uses.
func SignChallenge(privateKeyBase64 string, challenge string) (string, error) {
	privPem, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		privPem = []byte(privateKeyBase64)
	}

	block, _ := pem.Decode(privPem)
	if block == nil {
		return "", errors.New("failed to decode PEM block containing private key")
	}

	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	hash := sha256.Sum256([]byte(challenge))
	sig, err := ecdsa.SignASN1(rand.Reader, priv, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}
