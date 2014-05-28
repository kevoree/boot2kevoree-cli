package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	toml "github.com/BurntSushi/toml"
	vbx "github.com/boot2docker/boot2docker-cli/virtualbox"
	flag "github.com/ogier/pflag"
)

// boot2kevoree config.
var B2D struct {
	// NOTE: separate sections with blank lines so gofmt doesn't change
	// indentation all the time.

	// Gereral flags.
	Verbose bool
	VBM     string

	// basic config
	SSH      string // SSH client executable
	SSHGen   string // SSH keygen executable
	SSHKey   string // SSH key to send to the vm
	VM       string // virtual machine name
	Dir      string // boot2kevoree directory
	ISO      string // boot2kevoree ISO image path
	VMDK     string // base VMDK to use as persistent disk
	DiskSize uint   // VM disk image size (MB)
	Memory   uint   // VM memory size (MB)

	// NAT network: port forwarding
	SSHPort    uint16 // host SSH port (forward to port 22 in VM)
	DockerPort uint16 // host Docker port (forward to port 4243 in VM)
	WebGuiPort uint16 // host Kevoree port (forward to port 8080 in VM)
	GroupPort uint16 // host KevoreeGroup port (forward to port 9000 in VM)
	EditorPort uint16 // host KevoreeEditor port (forward to port 3042 in VM)

	// host-only network
	HostIP      net.IP
	DHCPIP      net.IP
	NetMask     net.IPMask
	LowerIP     net.IP
	UpperIP     net.IP
	DHCPEnabled bool
}

var (
	// Pattern to parse a key=value line in config profile.
	reFlagLine = regexp.MustCompile(`^\s*(\w+)\s*=\s*([^#;]+)`)
)

func getCfgDir(name string) (string, error) {
	if b2dDir := os.Getenv("BOOT2KEVOREE_DIR"); b2dDir != "" {
		return b2dDir, nil
	}

	dir := ""

	// *nix and MSYS Windows
	if dir = os.Getenv("HOME"); dir == "" {
		// Windows (if not running under MSYS)
		dir = os.Getenv("USERPROFILE")
	}
	if _, err := os.Stat(dir); err != nil {
		return "", err
	}
	dir = filepath.Join(dir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func getCfgFilename(dir string) string {
	filename := os.Getenv("BOOT2KEVOREE_PROFILE")
	if filename == "" {
		filename = filepath.Join(dir, "profile")
	}
	return filename
}

// Write configuration set by the combination of profile and flags
//    Should result in a format that can be piped into a profile file
func printConfig() string {
	var buf bytes.Buffer
	e := toml.NewEncoder(&buf)
	err := e.Encode(B2D)
	if err != nil {
		return ""
	}
	return buf.String()
}

// Read configuration from both profile and flags. Flags override profile.
func config() (*flag.FlagSet, error) {
	dir, err := getCfgDir(".boot2kevoree")
	if err != nil {
		return nil, fmt.Errorf("failed to get boot2kevoree directory: %s", err)
	}

	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.Usage = func() { usageLong(flags) }

	flags.StringVar(&B2D.VM, "vm", "boot2kevoree-vm", "virtual machine name.")
	// removed for now, requires re-parsing a new config file which is too messy
	//flags.StringVarP(&B2D.Dir, "dir", "d", dir, "boot2kevoree config directory.")
	B2D.Dir = dir
	flags.StringVar(&B2D.ISO, "iso", filepath.Join(dir, "boot2kevoree.iso"), "path to boot2kevoree ISO image.")
	flags.StringVar(&B2D.VMDK, "basevmdk", "", "Path to VMDK to use as base for persistent partition")
	vbm := "VBoxManage"
	if p := os.Getenv("VBOX_INSTALL_PATH"); p != "" && runtime.GOOS == "windows" {
		vbm = filepath.Join(p, "VBoxManage.exe")
	}
	flags.StringVar(&B2D.VBM, "vbm", vbm, "path to VirtualBox management utility.")
	flags.BoolVarP(&B2D.Verbose, "verbose", "v", false, "display verbose command invocations.")
	flags.StringVar(&B2D.SSH, "ssh", "ssh", "path to SSH client utility.")
	flags.StringVar(&B2D.SSHGen, "ssh-keygen", "ssh-keygen", "path to ssh-keygen utility.")

	sshdir, _ := getCfgDir(".ssh")
	flags.StringVar(&B2D.SSHKey, "sshkey", filepath.Join(sshdir, "id_boot2kevoree"), "path to SSH key to use.")
	flags.UintVarP(&B2D.DiskSize, "disksize", "s", 20000, "boot2kevoree disk image size (in MB).")
	flags.UintVarP(&B2D.Memory, "memory", "m", 1024, "virtual machine memory size (in MB).")
	flags.Uint16Var(&B2D.SSHPort, "sshport", 2022, "host SSH port (forward to port 22 in VM).")
	flags.Uint16Var(&B2D.DockerPort, "dockerport", 4243, "host Docker port (forward to port 4243 in VM).")
	flags.Uint16Var(&B2D.WebGuiPort, "kevguiport", 8080, "host KevoreeGui port (forward to port 8080 in VM).")
	flags.Uint16Var(&B2D.GroupPort, "kevgroup", 9000, "host KevoreeGroup port (forward to port 9000 in VM).")
	flags.Uint16Var(&B2D.EditorPort, "keveditorport", 3042, "host Kevoree Editor port (forward to port 3042 in VM).")


	flags.IPVar(&B2D.HostIP, "hostip", net.ParseIP("192.168.59.3"), "VirtualBox host-only network IP address.")
	flags.IPMaskVar(&B2D.NetMask, "netmask", flag.ParseIPv4Mask("255.255.255.0"), "VirtualBox host-only network mask.")
	flags.BoolVar(&B2D.DHCPEnabled, "dhcp", true, "enable VirtualBox host-only network DHCP.")
	flags.IPVar(&B2D.DHCPIP, "dhcpip", net.ParseIP("192.168.59.99"), "VirtualBox host-only network DHCP server address.")
	flags.IPVar(&B2D.LowerIP, "lowerip", net.ParseIP("192.168.59.103"), "VirtualBox host-only network DHCP lower bound.")
	flags.IPVar(&B2D.UpperIP, "upperip", net.ParseIP("192.168.59.254"), "VirtualBox host-only network DHCP upper bound.")

	// Set the defaults
	if err := flags.Parse([]string{}); err != nil {
		return nil, err
	}
	// Over-ride from the profile file
	filename := getCfgFilename(B2D.Dir)
	if _, err := os.Lstat(filename); err == nil {
		if _, err := toml.DecodeFile(filename, &B2D); err != nil {
			return nil, err
		}
	}

	// for cmd==ssh only:
	// only pass the params up to and including the `ssh` command - after that,
	// there might be other -flags that are destined for the ssh cmd
	i := 1
	for i < len(os.Args) && os.Args[i-1] != "ssh" {
		i++
	}
	// Command-line overrides profile config.
	if err := flags.Parse(os.Args[1:i]); err != nil {
		return nil, err
	}

	vbx.Verbose = B2D.Verbose
	vbx.VBM = B2D.VBM
	return flags, nil
}

func usageShort() {
	errf("Usage: %s [<options>] {editor|gui|help|init|up|ssh|save|down|poweroff|reset|restart|config|status|info|delete|download|version} [<args>]\n", os.Args[0])

}

func usageLong(flags *flag.FlagSet) {
	// NOTE: the help message uses spaces, not tabs for indentation!
	errf(`Usage: %s [<options>] <command> [<args>]

boot2kevoree management utility.

Commands:
    init                    Create a new boot2kevoree VM.
    up|start|boot           Start VM from any states.
    ssh [ssh-command]       Login to VM via SSH.
    editor                  Open the Kevoree editor.
    gui                     Open the default Kevoree cloud GUI in your browser.
    save|suspend            Suspend VM and save state to disk.
    down|stop|halt          Gracefully shutdown the VM.
    restart                 Gracefully reboot the VM.
    poweroff                Forcefully power off the VM (might corrupt disk image).
    reset                   Forcefully power cycle the VM (might corrupt disk image).
    delete                  Delete boot2kevoree VM and its disk image.
    config|cfg              Show selected profile file settings.
    info                    Display detailed information of VM.
    status                  Display current state of VM.
    download                Download boot2kevoree ISO image.
    version                 Display version information.

Options:
`, os.Args[0])
	flags.PrintDefaults()
}
