package v1

/*
func TLSVersionFromString(s string) (config.TLSVersion, error) {
	if s == "" {
		return 0, nil
	}
	if v, ok := config.TLSVersions[s]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("unknown TLS version: %s", s)
}

func (c *TLS) ToPrometheusConfig() (*config.TLSConfig, error) {
	var errs []error
	minVersion, err := TLSVersionFromString(c.MinVersion)
	if err != nil {
		errs = append(errs, fmt.Errorf("unable to convert TLS min version: %w", err))
	}
	maxVersion, err := TLSVersionFromString(c.MaxVersion)
	if err != nil {
		errs = append(errs, fmt.Errorf("unable to convert TLS min version: %w", err))
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return &config.TLSConfig{
		InsecureSkipVerify: c.InsecureSkipVerify,
		ServerName:         c.ServerName,
		MinVersion:         minVersion,
		MaxVersion:         maxVersion,
	}, nil
}

func (c *ProxyConfig) ToPrometheusConfig() (config.URL, error) {
	proxyURL, err := url.Parse(c.ProxyURL)
	if err != nil {
		return config.URL{}, fmt.Errorf("invalid proxy URL: %w", err)
	}
	// Marshalling the config will redact the password, so we don't support those.
	// It's not a good idea anyway and we will later support basic auth based on secrets to
	// cover the general use case.
	if _, ok := proxyURL.User.Password(); ok {
		return config.URL{}, errors.New("passwords encoded in URLs are not supported")
	}
	// Initialize from default as encode/decode does not work correctly with the type definition.
	return config.URL{URL: proxyURL}, nil
}

func (c *HTTPClientConfig) ToPrometheusConfig() (config.HTTPClientConfig, error) {
	var errs []error
	// Copy default config.
	clientConfig := config.DefaultHTTPClientConfig
	if c.TLS != nil {
		tlsConfig, err := c.TLS.ToPrometheusConfig()
		if err != nil {
			errs = append(errs, err)
		} else {
			clientConfig.TLSConfig = *tlsConfig
		}
	}
	if c.ProxyConfig.ProxyURL != "" {
		proxyConfig, err := c.ProxyConfig.ToPrometheusConfig()
		if err != nil {
			errs = append(errs, err)
		} else {
			clientConfig.ProxyURL = proxyConfig
		}
	}
	return clientConfig, errors.Join(errs...)
}
*/
