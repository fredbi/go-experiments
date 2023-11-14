// Package injected defines a interface to inject dependencies
// into all modules.
//
// Currently supported dependencies are:
// * a logger factory distributing context-aware loggers (runtime.Logger())
// * a config to spread a central *viper.Viper configuration registry (runtime.Config())
// * a database connection (runtime.DB())
// * an interface to persistent repositories (runtime.Repos())
package injected
