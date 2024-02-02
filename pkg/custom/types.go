package custom

type Connection struct {
	Url  string `mapstructure:url;json:url`
	Port string `mapstructure:port;json:port`
}
type CustomAnalyzer struct {
	Name       string     `mapstructure:name;json:name`
	Connection Connection `mapstructure:connection;json:connection`
}
