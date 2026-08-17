package configs

import "time"

type OrderConfig struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"password"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`
	Context struct {
		CreateOrderTimeOut time.Duration `yaml:"create_order_timeout"`
	} `yaml:"context"`
}
