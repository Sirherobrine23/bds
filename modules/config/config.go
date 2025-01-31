package config

// Config loaded
var ConfigProvider Config = &iniConfigProvider{disableSaving: true, file: "config.ini"}

func LoadConfig(file string) error {
	err := error(nil)
	ConfigProvider, err = NewConfigProviderFromFile(file)
	return err
}
