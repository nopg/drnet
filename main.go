package main

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
	"github.com/urfave/cli/v2" // imports as package "cli"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

type DCDR struct {
	DC []string `yaml:"DC"`
	DR []string `yaml:"DR"`
}

func getDevices() DCDR {
	yamlFile, err := os.ReadFile("dc-dr-devices.yaml")
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}
	var devices DCDR

	if err := yaml.Unmarshal(yamlFile, &devices); err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}
	fmt.Println("DC Devices: ")
	for _, device_ip := range devices.DC {
		fmt.Println(device_ip)
	}

	return devices
}

func getpass() string {
	fmt.Printf("Password: ")
	bytepw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("Error with password input.")
		os.Exit(1)
	}
	pass := string(bytepw)

	return pass
}

// Login to environment
func login(c *cli.Context) error {
	pass := getpass()

	//blah := []string{"diffie-hellman-group-exchange-sha1", "diffie-hellman-group1-sha1"}
	//options.WithStandardTransportExtraKexs(blah),

	p, err := platform.NewPlatform(
		"cisco_iosxe",
		"10.254.254.1",
		options.WithAuthNoStrictKey(),
		options.WithAuthUsername("admin"),
		options.WithAuthPassword(pass),
		options.WithSSHConfigFile("~/.ssh/config"),
	)
	if err != nil {
		fmt.Printf("Failed to create platform; error: %+v\n", err)
		return nil
	}
	d, err := p.GetNetworkDriver()
	if err != nil {
		fmt.Printf("Failed to fetch network driver from the platform; error: %+v\n", err)
		return nil
	}

	err = d.Open()
	if err != nil {
		fmt.Printf("Failed to open driver; error: %+v\n", err)
		return nil
	}
	defer d.Close()

	r, err := d.SendCommand("show version")
	if err != nil {
		fmt.Printf("Failed to send command; error: %+v\n", err)
		return nil
	}

	fmt.Printf("Sent command '%s', output received (SendCommand):\n %s\n\n\n", r.Input, r.Result)

	return nil
}

// Normal Operations
func normal(c *cli.Context) error {
	vlan := c.String("vlan")
	fmt.Printf("normal-vlan= %s\n", vlan)

	return nil
}

// Failback Operations
func failback(c *cli.Context) error {
	vlan := c.String("vlan")
	fmt.Printf("failback-vlan= %s \n", vlan)
	fmt.Print(vlan)

	_ = getDevices()

	return nil
}

// CLI Options
func main() {
	app := &cli.App{
		Name:  "drnet",
		Usage: "DR (Disaster Recovery) Network Helper Tool shut/no shut SVI's",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "vlan",
				Aliases:  []string{"v"},
				Usage:    "Interface Vlan Number(s?) to Migrate",
				Required: true,
			},
		},
		Commands: []*cli.Command{
			{
				Before: login,
				Name:   "normal",
				Usage:  "Normal Operations",
				Action: normal,
			},
			{
				//Before: login,
				Name:   "failback",
				Usage:  "Fail back from DR to DC",
				Action: failback,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
