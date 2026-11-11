package storage

import (
	"github.com/google/uuid"
	"golang.org/x/sys/windows/registry"
)

// GetMachineHWID returns the Windows MachineGuid from the registry or generates a fallback UUID.
func GetMachineHWID() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		guid, _, err := k.GetStringValue("MachineGuid")
		if err == nil && guid != "" {
			return guid
		}
	}
	return uuid.New().String()
}
