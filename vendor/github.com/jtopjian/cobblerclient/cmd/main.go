package main

import (
	"fmt"
	"net/http"

	cobbler "github.com/jtopjian/cobblerclient"
)

var config = cobbler.ClientConfig{
	Url:      "http://localhost:25151",
	Username: "cobbler",
	Password: "password",
}

func main() {
	c := cobbler.NewClient(http.DefaultClient, config)
	_, err := c.Login()
	if err != nil {
		fmt.Printf("Error logging in: %s\n", err)
	}

	fmt.Printf("Token: %s\n", c.Token)

	d := cobbler.Distro{
		Name:      "Test",
		Breed:     "Ubuntu",
		OSVersion: "trusty",
		Arch:      "x86_64",
		Kernel:    "/var/www/cobbler/ks_mirror/Ubuntu-14.04/install/netboot/ubuntu-installer/amd64/linux",
		Initrd:    "/var/www/cobbler/ks_mirror/Ubuntu-14.04/install/netboot/ubuntu-installer/amd64/initrd.gz",
	}

	fmt.Println("Creating a Distro")
	newDistro, err := c.CreateDistro(d)
	if err != nil {
		fmt.Printf("Error creating distro: %s\n", err)
	}

	fmt.Printf("New Distro: %+v\n", newDistro)

	if newDistro.Name != "Test" {
		fmt.Println("Distro name does not match.")
	}

	fmt.Println("Updating Distro")
	newDistro.Comment = "Update Test"
	if err := c.UpdateDistro(newDistro); err != nil {
		fmt.Printf("Error creating distro: %s\n", err)
	}

	fmt.Println("Creating a Profile")
	p := cobbler.Profile{
		Name:   "Testy",
		Distro: "Test",
	}

	newProfile, err := c.CreateProfile(p)
	if err != nil {
		fmt.Printf("Error creating profile: %s\n", err)
	}

	fmt.Printf("New Profile: %+v\n", newProfile)

	fmt.Println("Creating a System")
	eth0 := cobbler.Interface{
		MACAddress:    "aa:bb:cc:dd:ee:ff",
		Static:        true,
		InterfaceType: "bridge",
	}

	eth1 := cobbler.Interface{
		MACAddress: "aa:bb:cc:dd:ee:fa",
		Static:     true,
		Management: true,
	}

	s := cobbler.System{
		Comment:     "WTF",
		Name:        "Foobar",
		Profile:     "Testy",
		NameServers: []string{"8.8.8.8", "1.1.1.1"},
		PowerID:     "foo",
	}

	newSystem, err := c.CreateSystem(s)
	if err != nil {
		fmt.Printf("Error creating system: %s\n", err)
	}

	fmt.Printf("New System: %+v\n", newSystem)

	fmt.Println("Adding NIC to System")
	if err := newSystem.CreateInterface("eth0", eth0); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Adding second NIC to System")
	if err := newSystem.CreateInterface("eth1", eth1); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Syncing the cobbler server")
	if err := c.Sync(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Getting system")
	s2, err := c.GetSystem("Foobar")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Verifying NIC data")
	interfaces, err := s2.GetInterfaces()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n\n", interfaces)

	iface, err := s2.GetInterface("eth0")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("eth0:\n%+v\n\n", iface)

	iface, err = s2.GetInterface("eth1")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("eth1:\n%+v\n\n", iface)

	fmt.Println("Deleting Interface")
	err = s2.DeleteInterface("eth0")
	if err != nil {
		fmt.Println(err)
	}

	s2, err = c.GetSystem("Foobar")
	if err != nil {
		fmt.Println(err)
	}

	interfaces, err = s2.GetInterfaces()
	if err != nil {
		fmt.Println(err)
	}
	if len(interfaces) != 1 {
		fmt.Println("Error deleting interface eth1")
		fmt.Printf("%+v\n", interfaces)
	}

	fmt.Println("Deleting System")
	err = c.DeleteSystem("Foobar")
	if err != nil {
		fmt.Printf("Error deleting system: %s\n", err)
	}

	fmt.Println("Deleting Profile")
	err = c.DeleteProfile("Testy")
	if err != nil {
		fmt.Printf("Error deleting profile: %s\n", err)
	}

	fmt.Println("Deleting Distro")
	err = c.DeleteDistro("Test")
	if err != nil {
		fmt.Printf("Error deleting distro: %s\n", err)
	}

	fmt.Println("Creating a Snippet")
	snippet := cobbler.Snippet{
		Name: "/var/lib/cobbler/snippets/some-snippet",
		Body: "sample content",
	}

	err = c.CreateSnippet(snippet)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Deleting a Snippet")
	if err := c.DeleteSnippet("/var/lib/cobbler/snippets/some-snippet"); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Creating a Kickstart")
	ks := cobbler.KickstartFile{
		Name: "/var/lib/cobbler/kickstarts/foo.ks",
		Body: "sample content",
	}

	err = c.CreateKickstartFile(ks)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Deleting a Kickstart")
	if err := c.DeleteKickstartFile("/var/lib/cobbler/kickstarts/foo.ks"); err != nil {
		fmt.Println(err)
	}

}
