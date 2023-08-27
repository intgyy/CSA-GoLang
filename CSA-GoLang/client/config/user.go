package config

type ServerConfig struct {
	Name       string       `mapstructure:"name" json:"name"`
	Port       int          `mapstructure:"port" json:"port"`
	UserInfo   UserConfig   `mapstructure:"user_srv" json:"user_srv"`
	ConsulInfo ConsulConfig `mapstructure:"consul" json:"consul"`
}
type UserConfig struct {
	Name string `mapstructure:"name" json:"name"`
}
type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}
type NacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      uint64 `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	User      string `mapstructure:"user"`
	Password  string `mapstructure:"password"`
	DataId    string `mapstructure:"dataid"`
	Group     string `mapstructure:"group"`
}
