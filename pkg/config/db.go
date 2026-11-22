package config

import (
	"fmt"
	"net/url"
)

type DatabaseConfiguration struct {
	Driver   string
	Dbname   string
	Username string
	Password string
	Host     string
	Port     string
	LogMode  bool
	SslMode  string // master SSL mode, e.g. disable, require
	// Replica (read-only); if empty, master values are used
	ReplicaDbname   string
	ReplicaUsername string
	ReplicaPassword string
	ReplicaHost     string
	ReplicaPort     string
	ReplicaSslMode  string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // minutes
	ConnMaxIdleTime int // minutes
}

// DbConfiguration returns master and replica DSN from loaded config.
// Call after SetupConfig(). Returns empty strings if config not loaded.
func DbConfiguration() (masterDSN, replicaDSN string) {
	c := Get()
	if c == nil {
		return "", ""
	}
	d := &c.Database
	masterDSN = postgresDSN(d.Host, d.Username, d.Password, d.Dbname, d.Port, d.SslMode)
	replicaDSN = postgresDSN(d.ReplicaHost, d.ReplicaUsername, d.ReplicaPassword, d.ReplicaDbname, d.ReplicaPort, d.ReplicaSslMode)
	return masterDSN, replicaDSN
}

func postgresDSN(host, user, password, dbname, port, sslmode string) string {
	u := &url.URL{
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%s", host, port),
		Path:     "/" + dbname,
		RawQuery: fmt.Sprintf("sslmode=%s", url.QueryEscape(sslmode)),
	}
	if user != "" {
		u.User = url.UserPassword(user, password)
	}
	return u.String()
}
