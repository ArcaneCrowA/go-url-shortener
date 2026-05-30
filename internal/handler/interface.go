package handler

type Service interface {
	Shorten(site string) (string, error)
	Reroute(code string) (string, error)
}
