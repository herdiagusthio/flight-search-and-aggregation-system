package middleware

// RecoveryConfig holds configuration for the recovery middleware.
type RecoveryConfig struct {
	StackSize         int
	DisableStackAll   bool
	DisablePrintStack bool
}
