package custom

type Connection struct {
	Url  string `json:"url"`
	Port string `json:"port"`
}
type CustomAnalyzer struct {
	Name       string     `json:"name"`
	Connection Connection `json:"connection"`
}
