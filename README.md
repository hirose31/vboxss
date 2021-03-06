vboxss
======

Utility to manage snapshots of VirtualBox easily

Description
-----------

[Sahara vagrant plugin](https://github.com/jedi4ever/sahara) provides sandbox mode using snapshot feature of VirtualBox. But sahara can manage only one snapshot.

We can manage many snapshots by `vboxmanage` but it is annoying.

This `vboxss` is the utility to manage snapshots of VirtualBox easily.

- Can specify by short VM name
    - `vm1` instead of long `vm1_default_1404895653615_55181`
- Can restore only one command
    - internally `poweroff` and `startvm` before, after `restore` snapshot

Usage
-----

### List of running VMs

```
$ vboxss list
vm1                 vm1_default_1404895653615_55181
vm2                 vm2_default_1404967162355_44921
```

### List of VM's snapshots

```
$ vboxss list vm1
List of the snapshots of vm1_default_1404895653615_55181
initial                           e718f597-22b4-4bef-adf6-239fff78215f
install-apps                      5337e220-0075-46a5-8aaa-901361f352df
before-apply-chef                 703071da-26c1-4504-8c81-4535de33c2f2
```

### Take snapshot

```
$ vboxss take vm1 provisioned
```

### Restore VM using snapshot

```
$ vboxss restore vm1 before-apply-chef
```

NOTICE: `restore` subcommand performs `poweroff` and `startvm` before/after `restore`.

### Delete snapshot

```
$ vboxss delete vm1 provisioned
```

Installation
------------

### Just want the binary?

Go to the [releases page](https://github.com/hirose31/vboxss/releases), find the version you want, and download the zip file. Unpack the zip file, and put the binary to somewhere you want (on UNIX-y systems, /usr/local/bin or the like). Make sure it has execution bits turned on. Yes, it is a single binary! You can put it anywhere you want :)

_THIS IS THE RECOMMENDED WAY_

### go get

If you want to go the Go way (install in GOPATH/bin) and just want the command:

```
go get github.com/hirose31/vboxss
```


Authors
-------

* HIROSE Masaaki (hirose31)

