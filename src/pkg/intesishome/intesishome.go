package intesishome

import (
	"fmt"
	"log"
	"os"

	"github.com/nullify005/service-intesis/pkg/intesishome/cloud"
	"github.com/nullify005/service-intesis/pkg/intesishome/command"
)

type IntesisHome struct {
	username string
	password string
	cloud    *cloud.Cloud
	command  *command.Command
	logger   *log.Logger
}

type Option func(*IntesisHome)

func WithLogger(l *log.Logger) Option {
	return func(i *IntesisHome) {
		i.logger = l
	}
}

func WithCloud(c *cloud.Cloud) Option {
	return func(i *IntesisHome) {
		i.cloud = c
	}
}

func WithCommand(c *command.Command) Option {
	return func(i *IntesisHome) {
		i.command = c
	}
}

func New(username, password string, opts ...Option) *IntesisHome {
	logger := log.New(os.Stdout, "" /* prefix */, log.Ldate|log.Ltime|log.Lshortfile)
	i := &IntesisHome{
		username: username,
		password: password,
		logger:   logger,
		command:  command.New(command.WithLogger(logger)),
		cloud:    cloud.New(username, password, cloud.WithLogger(logger)),
	}
	for _, o := range opts {
		o(i)
	}
	return i
}

// contacts the Intesis Home API to obtain the status of a device
func (i *IntesisHome) Status(device int64) (state map[string]interface{}, err error) {
	state = make(map[string]interface{})
	response, err := i.cloud.Status()
	if err != nil {
		return
	}
	for _, s := range response.Status.Status {
		if s.DeviceID != device {
			continue
		}
		state[DecodeUid(s.UID)] = s.Value
	}
	return
}

func (i *IntesisHome) Set(device int64, key, value int) error {
	token, err := i.cloud.Token()
	if err != nil {
		return fmt.Errorf("unable to get command token. cause: %v", err)
	}
	i.command.Token(token)
	server, err := i.cloud.Command()
	if err != nil {
		return fmt.Errorf("unable to get command server. cause: %v", err)
	}
	err = i.command.Connect(server)
	if err != nil {
		return fmt.Errorf("unable to connect to command server. %v", err)
	}
	go i.command.Listen()
	err = i.command.Set(device, key, value)
	i.command.Close()
	return err
}

func (i *IntesisHome) Get(device int64, key string) error {
	return nil
}

// lists the devices confgured within Intesis Home
func (i *IntesisHome) Devices() (devices []Device, err error) {
	response, err := i.cloud.Status()
	if err != nil {
		return
	}
	for _, inst := range response.Config.Inst {
		for _, d := range inst.Devices {
			device := &Device{
				ID:             d.ID,
				Name:           d.Name,
				FamilyID:       d.FamilyID,
				ZoneID:         d.ZoneID,
				InstallationID: d.InstallationID,
				Order:          d.Order,
				Widgets:        d.Widgets,
			}
			devices = append(devices, *device)
		}
	}
	return
}

func (i *IntesisHome) HasDevice(device int64) (bool, error) {
	devices, err := i.Devices()
	if err != nil {
		return false, err
	}
	for _, d := range devices {
		if d.ID == fmt.Sprint(device) {
			return true, nil
		}
	}
	return false, nil
}
