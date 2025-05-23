package encrypt

import "testing"

var (
	DefaultKey       = "golang_test"
	PasswordToEncode = "Google#@128877"
)

func TestPassword(t *testing.T) {
	// Encrypt password
	passEncripted, err := Encrypt(DefaultKey, PasswordToEncode)
	if err != nil {
		t.Errorf("Cannot encrypt password: %s", err)
		return
	}
	t.Logf("Password encrypted: %s", passEncripted)

	// Descrypt password
	pass, err := Decrypt(DefaultKey, passEncripted)
	if err != nil {
		t.Errorf("Cannot decrypt password: %s", err)
		return
	} else if pass != PasswordToEncode {
		t.Errorf("passwords is not same:\n\t%s\n\t%s", PasswordToEncode, pass)
		return
	}

	// Fail descrypt
	pass, err = Decrypt(DefaultKey+"google", passEncripted)
	if err == nil || pass == PasswordToEncode {
		t.Errorf("password required return invalided:\n\t%s\n\t%s", PasswordToEncode, pass)
	}
}
