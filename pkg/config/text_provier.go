package config

var _ PasswordProvider = (*SimpleTextPasswordProvider)(nil)

type SimpleTextPasswordProvider struct {
	Password string
}

func (s SimpleTextPasswordProvider) GetPassword() (string, error) {
	return s.Password, nil
}
