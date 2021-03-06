package main

import (
	"github.com/julsemaan/garin/base"
	"gopkg.in/gcfg.v1"
	"reflect"
)

const DEFAULT_CONF_FILE = "garin.conf.defaults"

type Config struct {
	General struct {
		Log_level                string
		Parsing_concurrency      int
		Recording_threads        int
		Dont_record_destinations bool
	}
	Capture struct {
		Interface               string
		Unencrypted_ports       string
		Encrypted_ports         string
		Snaplen                 int
		Buffered_per_connection int
		Total_max_buffer        int
		Flush_after             string
	}
	Database struct {
		Type                  string
		Args                  string
		Debounce_destinations string
	}
}

func NewConfig(filename string) *Config {
	cfg := &Config{}
	err := gcfg.ReadFileInto(cfg, filename)
	if err != nil {
		base.Die("Failed to parse gcfg", err)
	}
	return cfg
}

func BuildConfig(filename string) *Config {
	default_cfg := NewConfig(DEFAULT_CONF_FILE)
	cfg := NewConfig(filename)

	reflect_default_cfg := reflect.ValueOf(default_cfg).Elem()
	reflect_cfg := reflect.ValueOf(cfg).Elem()
	for i := 0; i < reflect_default_cfg.NumField(); i++ {
		default_cfg_section := reflect_default_cfg.Field(i)
		cfg_section := reflect_cfg.Field(i)

		for j := 0; j < default_cfg_section.NumField(); j++ {
			default_field := default_cfg_section.Field(j)
			field := cfg_section.Field(j)
			if reflect.Value(field).Interface() != reflect.Zero(field.Type()).Interface() {
				default_field.Set(reflect.Value(field))
			} else {
				//fmt.Println("Not overriding default value for field %s - %s", reflect_default_cfg.Type().Field(i).Name, default_cfg_section.Type().Field(j).Name)
			}
		}
	}
	//fmt.Println("Starting using configuration : ", spew.Sdump(default_cfg))
	return default_cfg
}
