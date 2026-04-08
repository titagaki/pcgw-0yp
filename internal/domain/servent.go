package domain

type Servent struct {
	ID          int64
	Name        string
	Description string
	Hostname    string
	Port        int
	AuthID      string
	Passwd      string
	Priority    int
	MaxChannels int
	Enabled     bool
	Agent       string
	YellowPages string
}
