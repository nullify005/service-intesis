package intesishome

import (
	"fmt"
	"strings"
)

// representation of an Intesis Home device
type Device struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	FamilyID       int    `json:"familyId"`
	ModelID        int    `json:"modelId"`
	InstallationID int    `json:"installationId"`
	ZoneID         int    `json:"zoneId"`
	Order          int    `json:"order"`
	Widgets        []int  `json:"widgets"`
}

// emits a string of the device
func (d *Device) String() (s string) {
	s = fmt.Sprintf("device id: %v name: %v family: %v model: %v capabilities [%v]", d.ID, d.Name, d.FamilyID, d.ModelID, strings.Join(capabilities(d.Widgets), ","))
	return
}

// maps all the capabilities of the device
func capabilities(widgets []int) (caps []string) {
	for _, uid := range widgets {
		caps = append(caps, DecodeUid(uid))
	}
	return
}
