package httpservers

func getCookiePath() string {
	if cfg.Subpath == "" {
		return "/"
	}
	return normaliseSubpath(cfg.Subpath)
}
