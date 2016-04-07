package config

type UnikConfig struct {
	Config struct {
		Providers struct {
			Aws struct {
				AwsAccessKeyID    string `yaml:"aws_access_key_id"`
				AwsSecretAcessKey string `yaml:"aws_secret_acess_key"`
				Region            string `yaml:"region"`
			} `yaml:"aws"`
			Vsphere struct {
				VsphereUser     string `yaml:"vsphere_user"`
				VspherePassword string `yaml:"vsphere_password"`
				VsphereURL      string `yaml:"vsphere_url"`
			} `yaml:"vsphere"`
		} `yaml:"providers"`
	} `yaml:"config"`
}
