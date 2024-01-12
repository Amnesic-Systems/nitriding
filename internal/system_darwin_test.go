package internal

import "testing"

func TestNetworking(t *testing.T) {
	assertEqual(t, configureLoIface(), nil)
	assertEqual(t, configureTunIface(), nil)
	assertEqual(t, writeResolvconf(), nil)
}
