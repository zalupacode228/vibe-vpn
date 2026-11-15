package storage

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/sys/windows/registry"
)

// GetMachineHWID returns a salted SHA-256 hash of the Windows MachineGuid formatted as a UUID,
// preventing direct disclosure of the underlying system registry GUID.
func GetMachineHWID() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		guid, _, err := k.GetStringValue("MachineGuid")
		if err == nil && guid != "" {
			hash := sha256.Sum256([]byte("vibe_hwid_salt_" + guid))
			return fmt.Sprintf("%x-%x-%x-%x-%x", hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
		}
	}
	return uuid.New().String()
}
