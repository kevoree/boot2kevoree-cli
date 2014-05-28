package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	vbx "github.com/boot2docker/boot2docker-cli/virtualbox"
)

// TODO: delete the hostonlyif and dhcpserver when we delete the vm! (need to
// reference count to make sure there are no other vms relying on them)

// Get or create the hostonly network interface
func getHostOnlyNetworkInterface() (string, error) {
	// Check if the interface/dhcp exists.
	nets, err := vbx.HostonlyNets()
	if err != nil {
		return "", err
	}

	dhcps, err := vbx.DHCPs()
	if err != nil {
		return "", err
	}

	for _, n := range nets {
		if dhcp, ok := dhcps[n.NetworkName]; ok {
			if dhcp.IPv4.IP.Equal(B2D.DHCPIP) &&
				dhcp.IPv4.Mask.String() == B2D.NetMask.String() &&
				dhcp.LowerIP.Equal(B2D.LowerIP) &&
				dhcp.UpperIP.Equal(B2D.UpperIP) &&
				dhcp.Enabled == B2D.DHCPEnabled {
				if B2D.Verbose {
					logf("Reusing existing host-only network interface %q", n.Name)
				}
				return n.Name, nil
			}
		}
	}

	// No existing host-only interface found. Create a new one.
	if B2D.Verbose {
		logf("Creating a new host-only network interface...")
	}
	hostonlyNet, err := vbx.CreateHostonlyNet()
	if err != nil {
		return "", err
	}
	hostonlyNet.IPv4.IP = B2D.HostIP
	hostonlyNet.IPv4.Mask = B2D.NetMask
	if err := hostonlyNet.Config(); err != nil {
		return "", err
	}

	// Create and add a DHCP server to the host-only network
	dhcp := vbx.DHCP{}
	dhcp.IPv4.IP = B2D.DHCPIP
	dhcp.IPv4.Mask = B2D.NetMask
	dhcp.LowerIP = B2D.LowerIP
	dhcp.UpperIP = B2D.UpperIP
	dhcp.Enabled = true
	if err := vbx.AddHostonlyDHCP(hostonlyNet.Name, dhcp); err != nil {
		return "", err
	}
	return hostonlyNet.Name, nil
}

// Copy disk image from given source path to destination
func copyDiskImage(dst, src string) (err error) {
	// Open source disk image
	srcImg, err := os.Open(src)
	if err != nil {
		return err
	}
	closeSrcImg := func() {
		if ee := srcImg.Close(); ee != nil {
			err = ee
		}
	}
	defer closeSrcImg()
	dstImg, err := os.Create(dst)
	if err != nil {
		return err
	}
	closeDstImg := func() {
		if ee := dstImg.Close(); ee != nil {
			err = ee
		}
	}
	defer closeDstImg()
	_, err = io.Copy(dstImg, srcImg)
	return err
}

// Make a boot2kevoree VM disk image with the given size (in MB).
func makeDiskImage(dest string, size uint, initialBytes []byte) error {
	// Create the dest dir.
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	// Fill in the magic string so boot2kevoree VM will detect this and format
	// the disk upon first boot.
	raw := bytes.NewReader(initialBytes)
	return vbx.MakeDiskImage(dest, size, raw)
}
