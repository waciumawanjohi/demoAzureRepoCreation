package env

import (
	"fmt"
	"net/url"
	"os"
	"time"

	b64 "encoding/base64"
)

type vvar struct {
	key           string
	defaultValue  string
	allowedValues []string
	required      bool
	value         string
	fallback      *vvar
}

func Var(key string) *vvar {
	return &vvar{
		key:           key,
		allowedValues: []string{},
	}
}

func (v *vvar) AllowedValues(vals ...string) *vvar {
	v.allowedValues = append(v.allowedValues, vals...)
	return v
}

func (v *vvar) Key() string {
	return v.key
}

func (v *vvar) Required() *vvar {
	v.required = true
	return v
}

func (v *vvar) Default(defaultValue string) *vvar {
	v.defaultValue = defaultValue
	return v
}

func (v *vvar) FallsbackTo(fallback *vvar) *vvar {
	v.fallback = fallback
	return v
}

func (v *vvar) ResolveDuration() (time.Duration, error) {
	value, err := v.Resolve()
	if err != nil {
		return -1, err
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return -1, fmt.Errorf("parse duration '%s': %w", value, err)
	}

	return duration, nil
}

func (v *vvar) ResolveUrl() (*url.URL, error) {
	value, err := v.Resolve()
	if err != nil {
		return nil, err
	}

	fullUrl, err := url.Parse(value)
	if err != nil {
		return fullUrl, fmt.Errorf("parse url '%s': %w", value, err)
	}

	return fullUrl, nil
}

func (v *vvar) Resolve() (string, error) {
	val, err := v.resolve()
	if err != nil {
		return "", err
	}

	if len(v.allowedValues) > 0 {
		if !contains(v.allowedValues, val) {
			return "", fmt.Errorf("value '%s' is not part of "+
				"allowed values '%v'", val, v.allowedValues)
		}
	}

	return val, nil
}

func (v *vvar) resolve() (string, error) {
	value := os.Getenv(v.key)
	if value != "" {
		return value, nil
	}

	if v.required {
		return "", fmt.Errorf("env '%s' required but empty", v.key)
	}

	if v.fallback != nil {
		return v.fallback.Resolve()
	}

	return v.defaultValue, nil
}

func (v *vvar) MustResolve() string {
	value, err := v.Resolve()
	if err != nil {
		panic(fmt.Errorf("resolve: %w", err))
	}

	return value
}

func (v *vvar) MustResolveDuration() time.Duration {
	value, err := v.ResolveDuration()
	if err != nil {
		panic(err)
	}

	return value
}

func (v *vvar) MustResolveUrl() *url.URL {
	value, err := v.ResolveUrl()
	if err != nil {
		panic(err)
	}

	return value
}

func (v *vvar) MustResolveBase64() string {
	value, err := v.Resolve()

	if err != nil {
		panic(fmt.Errorf("resolve: %w", err))
	}

	sDec, err := b64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(fmt.Errorf("base64 decode: %w", err))
	}

	return string(sDec)
}

func contains(slice []string, val string) bool {
	for _, elementInSlice := range slice {
		if val == elementInSlice {
			return true
		}
	}

	return false
}
