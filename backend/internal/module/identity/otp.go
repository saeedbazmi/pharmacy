package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// SMSSender delivers an OTP. Implementations must never write the code to
// structured logs; the development sender may print it only when explicitly
// enabled.
type SMSSender interface {
	SendOTP(phone, code string) error
	Notify(phone, message string) error
}

type devSender struct {
	printCode bool
	out       io.Writer
}

// NewDevSender is used in development. The code is written to out only when
// printCode is true, never through slog.
func NewDevSender(printCode bool, out io.Writer) SMSSender {
	if out == nil {
		out = os.Stdout
	}
	return &devSender{printCode: printCode, out: out}
}

func (s *devSender) SendOTP(phone, code string) error {
	if s.printCode {
		_, _ = fmt.Fprintf(s.out, "otp.dev_preview phone=%s code=%s\n", MaskPhone(phone), code)
	}
	return nil
}

func (s *devSender) Notify(phone, message string) error {
	if s.printCode {
		_, _ = fmt.Fprintf(s.out, "alert.dev_preview phone=%s message=%s\n", MaskPhone(phone), message)
	}
	return nil
}

type unavailableSender struct{}

func NewUnavailableSender() SMSSender { return unavailableSender{} }

func (unavailableSender) SendOTP(string, string) error { return ErrSMSUnavailable }
func (unavailableSender) Notify(string, string) error  { return ErrSMSUnavailable }

func hashOTP(pepper, phone, code string) string {
	sum := sha256.Sum256([]byte(pepper + "\n" + phone + "\n" + code))
	return hex.EncodeToString(sum[:])
}

func compareOTP(pepper, phone, code, stored string) bool {
	got := hashOTP(pepper, phone, code)
	if len(got) != len(stored) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(stored)) == 1
}

func randomDigits(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = '0' + (b % 10)
	}
	return string(out), nil
}

func randomToken() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
