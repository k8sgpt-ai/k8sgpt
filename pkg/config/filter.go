package config

import (
	"github.com/spf13/viper"
)

func ListActiveFilters() []string {
	return viper.GetStringSlice("active_filters")
}

func SetActiveFilters(filters []string) {
	viper.Set("active_filters", filters)
}
