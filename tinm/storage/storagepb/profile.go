package storagepb

import (
    "google.golang.org/protobuf/encoding/protojson"
	"errors"
)

var (
	ErrIdRequired = errors.New("Id is required")
)

// ParseProfile parses bytes into a Profile.
func ParseProfile(data []byte) (*Profile, error) {
	profile := new(Profile)
	err := protojson.Unmarshal(data, profile)
	return profile, err
}

// AssertValid validates a Profile. Returns nil if there are no validation
// errors.
func (p *Profile) AssertValid() error {
	// Id is required
	if p.Id == "" {
		return ErrIdRequired
	}
	return nil
}

func (p *Profile) Copy() *Profile {
	if p == nil {
		return nil
	}

	cp := &Profile{
		Id:         p.Id,
		Name:       p.Name,
		IgnitionId: p.IgnitionId,
		CloudId:    p.CloudId,
		GenericId:  p.GenericId,
	}

	switch v := p.GetBootMode().(type) {
	case *Profile_Boot:
		cp.BootMode = &Profile_Boot{
			Boot: v.Boot.Copy(),
		}
	case *Profile_Chain:
		cp.BootMode = &Profile_Chain{
			Chain: v.Chain,
		}
	}

	return cp
}

func (b *NetBoot) Copy() *NetBoot {
	initrd := make([]string, len(b.Initrd))
	copy(initrd, b.Initrd)
	args := make([]string, len(b.Args))
	copy(args, b.Args)
	return &NetBoot{
		Kernel: b.Kernel,
		Initrd: initrd,
		Args:   args,
	}
}
