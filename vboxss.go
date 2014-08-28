package main

import(
	"fmt"
	"os"
	"flag"
	"log"
	"bytes"
	"os/exec"
	"io"
	"regexp"
)

type command struct {
	fs *flag.FlagSet
	fn func(args []string) error
	usage func()
}

type vm struct {
	vmname string
	short_vmname string
	uuid string
}

type snapshot struct {
	name string
	uuid string
}

func main() {
	commands := map[string]command{
		"list": listCmd(),
		"take": takeCmd(),
		"restore": restoreCmd(),
		"delete": deleteCmd(),
	}

	flag.Usage = func() {
		fmt.Println("Usage: vboxss <command> [options]")
		for name, cmd := range commands {
			if cmd.usage != nil {
				fmt.Print("\n")
				cmd.usage()
			} else {
				fmt.Printf("\n%s [options]:\n", name)
				cmd.fs.PrintDefaults()
			}
		}
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if cmd, ok := commands[args[0]]; !ok {
		log.Fatalf("Unknown command: %s", args[0])
	} else if err := cmd.fn(args[1:]); err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func runCommand(args []string) (out bytes.Buffer, oer bytes.Buffer, err error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &out
	cmd.Stderr = &oer
	err = cmd.Run()
	if err != nil {
		if oer.String() != "" {
			log.Println(oer.String())
		}
		return out, oer, err
	}

	return out, oer, nil
}

func normalize_vmname(vmname string) (string) {
	var candidate_vms []vm

	vms, err := retrieve_vms()
	if err != nil {
		return vmname
	}

	for _, vm := range vms {
		if vmname == vm.short_vmname || vmname == vm.vmname {
			candidate_vms = append(candidate_vms, vm)
		}
	}

	if len(candidate_vms) == 1 {
		return candidate_vms[0].vmname
	} else {
		log.Printf("Found several VMs for `%s`. You must specify not short vmname but long vmname.", vmname)
		for _, vm := range candidate_vms {
			log.Println(vm.vmname)
		}
		log.Fatal("Aborted")
		return ""
	}
}

func retrieve_vms() (vms []vm, err error) {
	cmd := []string{"vboxmanage", "list", "runningvms"}
	out, oer, err := runCommand(cmd)
	if err != nil {
		log.Println(oer.String())
		return nil, err
	}

	re_line, _ := regexp.Compile(`"([^"]+)"\s+{([^}]+)}`)
	re_vmname, _ := regexp.Compile(`^(.+)_default_[0-9_]+`)
	for {
		line, err := out.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		match := re_line.FindStringSubmatch(line)
		if match != nil {
			vmname := match[1]
			uuid := match[2]

			short_vmname := vmname
			match := re_vmname.FindStringSubmatch(vmname)
			if match != nil {
				short_vmname = match[1]
			}

			vms = append(vms, vm{vmname, short_vmname, uuid})
		} else {
			continue
		}
	}

	return vms, nil
}

func retrieve_snapshots(vmname string) (snapshots []snapshot, err error) {
	cmd := []string{"vboxmanage", "snapshot", vmname, "list"}
	out, oer, err := runCommand(cmd)
	if err != nil {
		if m, _ := regexp.MatchString(`This machine does not have any snapshots`, out.String()); m {
			return nil, nil
		} else {
			log.Println(oer.String())
			return nil, err
		}
	}

	re_line, _ := regexp.Compile(`Name:\s*(.+)\s\(UUID:\s*([^)]+)`)
	for {
		line, err := out.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		match := re_line.FindStringSubmatch(line)
		if match != nil {
			name := match[1]
			uuid := match[2]

			snapshots = append(snapshots, snapshot{name, uuid})
		} else {
			continue
		}
	}

	return snapshots, nil
}

func listCmd() command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Println("list\n  Print information about running VMs")
		fmt.Print("\n");
		fmt.Println("list <vmname>\n  Print information about snapshots of the VM")
		fs.PrintDefaults()
	}

	return command{
		fs,
		func(args []string) error {
			fs.Parse(args)

			if len(args) == 0 {
				return listVm()
			} else if len(args) == 1 {
				return listSnapshot(args)
			} else {
				return fmt.Errorf("args of list command is 0 or 1")
			}
		},
		fs.Usage,
	}
}

func listVm() (err error) {
	vms, err := retrieve_vms()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		fmt.Printf("%-18s %s\n", vm.short_vmname, vm.vmname)
	}

	return nil
}

func listSnapshot(args []string) (err error) {
	vmname := normalize_vmname(args[0])

	fmt.Printf("List of the snapshots of %s\n", vmname)

	snapshots, err := retrieve_snapshots(vmname)
	if err != nil {
		return err
	}
	if snapshots == nil {
		fmt.Println("No snapshot")
		return nil
	}

	for _, ss := range snapshots {
		fmt.Printf("%-39s %s\n", ss.name, ss.uuid)
	}

	return err
}

func takeCmd() command {
	fs := flag.NewFlagSet("take", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Println("take <vmname> <snapshot name>\n  Take the snapshot of the specified VM")
		fs.PrintDefaults()
	}

	return command{
		fs,
		func(args []string) error {
			fs.Parse(args)

			if len(args) == 2 {
				return takeSnapshot(args)
			} else {
				return fmt.Errorf("Invalid options")
			}
		},
		fs.Usage,
	}
}

func takeSnapshot(args []string) (err error) {
	vmname := normalize_vmname(args[0])
	ss_name := args[1]

	fmt.Printf("Take snapshot of %s as '%s'... ", vmname, ss_name)

	cmd := []string{"vboxmanage", "snapshot", vmname, "take", ss_name, "--pause"}
	_, oer, err := runCommand(cmd)
	if err != nil {
		fmt.Println("error!")
		log.Println(oer.String())
		return err
	}

	fmt.Println("done!")
	return nil
}

func deleteCmd() command {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Println("delete <vmname> <snapshot name>\n  Delete the snapshot of the specified VM")
		fs.PrintDefaults()
	}

	return command{
		fs,
		func(args []string) error {
			fs.Parse(args)

			if len(args) == 2 {
				return deleteSnapshot(args)
			} else {
				return fmt.Errorf("Invalid options")
			}
		},
		fs.Usage,
	}
}

// fixme accept uuid
func deleteSnapshot(args []string) (err error) {
	vmname := normalize_vmname(args[0])
	ss_name := args[1]

	fmt.Printf("Delete snapshot of %s named '%s'... ", vmname, ss_name)

	cmd := []string{"vboxmanage", "snapshot", vmname, "delete", ss_name}
	_, oer, err := runCommand(cmd)
	if err != nil {
		fmt.Println("error!")
		log.Println(oer.String())
		return err
	}

	fmt.Println("done!")
	return nil
}

func restoreCmd() command {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Println("restore <vmname> <snapshot name>\n  Restore VM from the specified snapshot")
		fs.PrintDefaults()
	}

	return command{
		fs,
		func(args []string) error {
			fs.Parse(args)

			if len(args) == 2 {
				return restoreSnapshot(args)
			} else {
				return fmt.Errorf("Invalid options")
			}
		},
		fs.Usage,
	}
}

// fixme accept uuid
func restoreSnapshot(args []string) (err error) {
	vmname := normalize_vmname(args[0])
	ss_name := args[1]

	fmt.Printf("Restore VM(%s) from the snapshot named '%s'... ", vmname, ss_name)

	cmds := [][]string{
		{"vboxmanage", "controlvm", vmname, "poweroff"},
		// fixme check snapsnot exists or not
		{"vboxmanage", "snapshot", vmname, "restore", ss_name},
		{"vboxmanage", "startvm", "--type", "headless", vmname},
	}

	for _, cmd := range cmds {
		_, oer, err := runCommand(cmd)
		if err != nil {
			fmt.Println("error!")
			log.Println(oer.String())
			return err
		}

	}

	fmt.Println("done!")
	return nil
}

