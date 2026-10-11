package g6enforcement_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"time"
)

// g6DevFixturePrivateKeyPEM signs test JWTs for the live deploy/g6 topology
// only. Not a secret — see deploy/g6/jwks/README.md. Its public half is
// deploy/g6/jwks/dev-fixture-jwks.json, which agentgateway's jwtAuth.jwks.file
// points at to verify these tokens' signatures. Deliberately duplicated in
// deploy/g6/probe-client/jwt.go (a separate Go module) — see that README for
// why this small fixture isn't factored into a shared module.
const g6DevFixturePrivateKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAtRCOLetnN3HQgwJONY8SyWoCwUwbq+2mBVvOaFXPJfSw1sJV
XWQ5sHKpxOTiEUbcQqRnhvIvG9bK7CrzjFffpfD+EWtI3pqo10ekQa9RP1iFGrCc
U9PRblyVWWaSAMzHxTiWq5fo3T0+PwT4HZPHindAPwJlhUF2U5HNFrDerp1VDUf8
uN4A8LtKurbBRsdeHmZTYQx+7EFnfQFO2zObAMVJid8mYjp5bDFW2QdBPzomZGbZ
I12fNII8dzyzcK2oyqYDTvNKmgbI6QFKymkPiHKeb+XJZnEOEk8jglIpTMx4KwBw
Qcr1UmwfKEC36ZWnZAN5LoFDNspmG7OOcew/VQIDAQABAoIBAAm9pI0RlPrOnvft
VGWBzhalkMvOf86aVKeXboIjADnpXufTZb1qw313LbZdZ5nIoGoDHjfVceNaF2nE
BcYzpk7KhM5NYYrF8q0ZFZLTDyGfAFxwXQrqBuJTVnoeffDhpbSDDtgFiddBtVi1
byhDnm48kLKJ8eVfDw1tm7wUxYqmEpXhrDOPfrBp4PoZ4QTTDy/K2tgTw/M1Scsg
NC+yVi2nYmpkxlbuuc9YTcrlutDAHWmEzTxT+HAUGT902pVPyDoSr3HoPmMGbS5T
5UDJ4Y+QLTsGK2hqKt382wi+orxIyjodTn4Or/VdxyFDLq5KE/M/RMnsLk1qdzH4
bfRJaSECgYEA8MeOsyEMsiM//7VimmpFOb83TGd8Izo8C/y4HU24GeUmpsdzjXOG
qRchVw4snTsHb4LHB0PBUhxIDjdlqfL0sD7fhieRTp8IDJS7OTBAUvDZKS69WSJz
ntmb5eqCFmpADHPsfVy4aHJMP76QUfOUTw4LF7HL8ZQli6tR9qv0vFsCgYEAwIKn
yv9G5WwE8UbtHEA0rWrRpZb/woB5QJW3XGDx/b9v/qhW9u5VqHmWTmSc/Cu72fJ5
ko8c/gzxuVF8r/ZKM9OXqOJH/VHcHfxmgPTBa1SsMiNz0TVlLD5c8qsb/6uGOOKY
LaagLdDFbXcMGQM6MPKSpF4Ed7DxsR+qtKP9gg8CgYEAziJaddrmjp+FC/sS1pYC
fATLZ9r0uQgDHlQWn+fIpEq9Q21f7Qqpj5ugzHHzGgzOOdZhZEPKfux9d8ZPgCbi
+vxoyuaXDRMzhenTO4umlhtiH1LHgkbva2BrinOxOVVvTfn0zgKSUcEArFYOIksB
fojMUFXD/ydQ2Xkra54doR8CgYAm0XHCNi12j4yDlnizbKLyoQp7KHKUJtHMWyQp
JYdGUnbj09ANZMuy+Cl9zz30f2EWtpUbH26KL9QCOVM6LCCUSMNZE5/OjdYj2cRV
loT1/pHmXk25TtoCzORzLlur90tOZyqmceX0txdIVmwDEyqFujQlnqup8u0ZeTgz
yqmQswKBgFZC2F1wWRBD1G8NE2WP9XpgYLwupkj3s6q6DxjqgvyP9IMAYpZ8NcRd
5rmakcUew4l2FtRilD4snicx2NaNzf0FKm3jeVukvm3OfUUfPfka5GhdtckiR4rX
BZm9PqX/kzvFRPV/rZo4kbM7kBei4IsbM25dcvQ49SMgow2R9StI
-----END RSA PRIVATE KEY-----`

const g6DevFixtureIssuer = "https://g6-dev-fixture.invalid/"
const g6DevFixtureAudience = "agentgate-g6"

func g6MustLoadFixtureKey() *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(g6DevFixturePrivateKeyPEM))
	if block == nil {
		panic("jwt_fixture_test: failed to decode dev fixture private key PEM")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic("jwt_fixture_test: failed to parse dev fixture private key: " + err.Error())
	}
	return key
}

func base64URLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// mintTestJWT signs a minimal RS256 JWT for the live deploy/g6 topology, with
// the given claims merged on top of standard iss/aud/exp/iat. Stdlib only
// (crypto/rsa) — deliberately not a JWT library dependency for ~20 lines of
// encoding.
func mintTestJWT(claims map[string]any) string {
	header := map[string]any{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "g6-dev-fixture-key-1",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		panic(err)
	}

	now := time.Now().UTC()
	payload := map[string]any{
		"iss": g6DevFixtureIssuer,
		"aud": g6DevFixtureAudience,
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
	}
	for k, v := range claims {
		payload[k] = v
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	signingInput := base64URLEncode(headerJSON) + "." + base64URLEncode(payloadJSON)

	digest := sha256.Sum256([]byte(signingInput))
	key := g6MustLoadFixtureKey()
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		panic("jwt_fixture_test: sign: " + err.Error())
	}

	return signingInput + "." + base64URLEncode(sig)
}

// jwtBearerHeader mints a token from claims and wraps it as the single
// header a *_LiveE2E test needs to pass to sendMCPRequest for identity.
func jwtBearerHeader(claims map[string]any) map[string]string {
	return map[string]string{"Authorization": "Bearer " + mintTestJWT(claims)}
}

// mergeHeaders combines a JWT auth header with any additional non-identity
// headers a scenario needs (e.g. x-tool-fingerprint), which agentgateway
// forwards as plain headers untouched — only identity comes from the JWT.
func mergeHeaders(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
