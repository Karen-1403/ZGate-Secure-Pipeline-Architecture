package config

// NOTE: Update these paths and credentials before deploying.

// Listening address for the proxy
const ListenAddr = "127.0.0.1:8443"

// TLS cert/key and CA; the repository README explains how to generate them
const ServerCertPath = "certs/server.crt"
const ServerKeyPath = "certs/server.key"
const CACertPath = "certs/ca.crt"

// MongoDB connection used by the proxy (the proxy will use this to connect to the real DB)
// This must be accessible only by the proxy. Clients use mTLS to authenticate to the proxy
const MongoURI = "mongodb://proxy_user:proxy_password@127.0.0.1:27017/?authSource=admin"

// A minimal in-memory user store for demo purposes. In production, use a secure user DB.
var Users = map[string]string{
	// username: plaintext password (demo only)
	"alice": "alice-pass",
	"bob":   "bob-pass",
}

// Simple authorization map: user -> allowed collections
var UserAllowedCollections = map[string][]string{
	"alice": {"orders", "users"},
	"bob":   {"orders"},
}
