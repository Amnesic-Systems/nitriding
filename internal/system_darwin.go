package internal

// Nitriding does not run on macOS but by implementing the following dummy
// functions, we can at least get it to compile.
func configureLoIface() error  { return nil }
func configureTunIface() error { return nil }
func writeResolvconf() error   { return nil }
func maybeSeedEntropy()        {}
