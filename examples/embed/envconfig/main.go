// Environment Config Example
// Demonstrates loading environment variables and config into Logos scripts

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/codetesla51/logos/logos"
)

func main() {
	// Set some example environment variables for demo
	os.Setenv("APP_NAME", "MyApp")
	os.Setenv("APP_ENV", "development")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATABASE_PORT", "5432")
	os.Setenv("DATABASE_NAME", "mydb")
	os.Setenv("API_KEY", "secret-key-12345")
	os.Setenv("MAX_CONNECTIONS", "100")

	// Create VM with limited permissions
	vm := logos.NewWithConfig(logos.SandboxConfig{
		AllowFileIO:  false,
		AllowNetwork: false,
		AllowShell:   false,
		AllowExit:    false,
	})

	// Register env functions
	vm.Register("env", func(args ...logos.Object) logos.Object {
		if len(args) < 1 {
			return &logos.Null{}
		}
		key := args[0].(*logos.String).Value
		value := os.Getenv(key)
		if value == "" {
			if len(args) > 1 {
				// Return default value if provided
				return args[1]
			}
			return &logos.Null{}
		}
		return &logos.String{Value: value}
	})

	vm.Register("envInt", func(args ...logos.Object) logos.Object {
		if len(args) < 1 {
			return &logos.Null{}
		}
		key := args[0].(*logos.String).Value
		value := os.Getenv(key)
		if value == "" {
			if len(args) > 1 {
				return args[1]
			}
			return &logos.Integer{Value: 0}
		}
		var intVal int64
		fmt.Sscanf(value, "%d", &intVal)
		return &logos.Integer{Value: intVal}
	})

	vm.Register("envBool", func(args ...logos.Object) logos.Object {
		if len(args) < 1 {
			return &logos.Bool{Value: false}
		}
		key := args[0].(*logos.String).Value
		value := strings.ToLower(os.Getenv(key))
		isTruthy := value == "true" || value == "1" || value == "yes" || value == "on"
		return &logos.Bool{Value: isTruthy}
	})

	vm.Register("envAll", func(args ...logos.Object) logos.Object {
		prefix := ""
		if len(args) > 0 {
			prefix = args[0].(*logos.String).Value
		}

		pairs := make(map[string]logos.Object)
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key, value := parts[0], parts[1]
				if prefix == "" || strings.HasPrefix(key, prefix) {
					pairs["STRING:"+key] = &logos.String{Value: value}
				}
			}
		}
		return &logos.Table{Pairs: pairs}
	})

	vm.Register("requireEnv", func(args ...logos.Object) logos.Object {
		if len(args) < 1 {
			return &logos.String{Value: "error: key required"}
		}
		key := args[0].(*logos.String).Value
		value := os.Getenv(key)
		if value == "" {
			return &logos.String{Value: "error: required env var not set: " + key}
		}
		return &logos.String{Value: value}
	})

	// Config script
	script := `
// Environment Configuration Script
print("=== Application Configuration ===")
print("")

// Basic env access
let appName = env("APP_NAME", "DefaultApp")
let appEnv = env("APP_ENV", "production")
let isDebug = envBool("APP_DEBUG")

print("App Name: " + appName)
print("Environment: " + appEnv)
print("Debug Mode: " + toStr(isDebug))
print("")

// Database config
print("--- Database Config ---")
let dbHost = env("DATABASE_HOST", "localhost")
let dbPort = envInt("DATABASE_PORT", 5432)
let dbName = env("DATABASE_NAME", "app")

print("Host: " + dbHost)
print("Port: " + toStr(dbPort))
print("Database: " + dbName)
print("")

// Check required env vars
print("--- Validating Required Vars ---")
let apiKey = requireEnv("API_KEY")
if startsWith(apiKey, "error:") {
    print(colorRed(apiKey))
} else {
    print(colorGreen("API_KEY is set"))
}
print("")

// Get all APP_ prefixed vars
print("--- All APP_* Variables ---")
let appVars = envAll("APP_")
for pair in appVars {
    print("  " + pair[0] + " = " + pair[1])
}
print("")

// Build connection string
let connStr = "postgres://" + dbHost + ":" + toStr(dbPort) + "/" + dbName
print("Connection String: " + connStr)
print("")

// Environment-specific behavior
print("--- Environment Behavior ---")
if appEnv == "development" {
    print(colorYellow("Running in DEVELOPMENT mode"))
    print("  - Verbose logging enabled")
    print("  - Hot reload active")
}
if appEnv == "production" {
    print(colorGreen("Running in PRODUCTION mode"))
    print("  - Optimizations enabled")
    print("  - Error reporting active")
}

if isDebug {
    print("")
    print(colorMagenta("DEBUG INFO:"))
    print("  Max Connections: " + toStr(envInt("MAX_CONNECTIONS", 10)))
}
`

	fmt.Println("Loading environment configuration...")
	fmt.Println("")

	err := vm.Run(script)
	if err != nil {
		fmt.Println("Script error:", err)
	}
}
