# Virtual Box Manager

This is a tool to easily manage virtual boxes. Goals are to:

* simplify usage of VirtualBox.
* have a set of machine images available that work out of the box.
* make cloning, starting, and deleting as fast as possible.


## Installation

Install [VirtualBox](http://download.virtualbox.org/virtualbox/4.3.6/VirtualBox-4.3.6-91406-OSX.dmg) and the vbox tool:
	go get github.com/dynport/dgtk/vbox

Now retrieve your local template engine.
	vbox get template ubuntu_precise.ova template

`ubuntu_precise.ova` is the image to be downloaded and `template` the name of the machine created (`template` is the
default in the `clone` command). The server can be provided using `--source` option.


## Usage

The following actions are available:

* `vbox list` will return the list of all available virtual machines (add `-r` to only show running machines).
* `vbox vm clone <vm>` will clone the template you just downloaded to a new machine with the `<vm>` name.
* `vbox vm start <vm>` will start the virtual machine with the given name.
* `vbox vm ssh into <vm>` will open an ssh connection to this machine (might take a few seconds as it tries to wait till
  the machine finished booting).

For more help run `vbox -h` or `vbox <command> -h`.


## Creating Images

Creating (or extending) images requires the following steps to be taken:

* Install or extend the image as wanted.
* Make sure the VirtualBox GuestAdditions are installed (for Linux remember to first install the `dkms` package).
* Remove the file `/etc/udev/rules.d/70-persistent-net.rules` so that the template's MAC is reset afterwards (otherwise
  networking will fail horribly).
* Shutdown the virtual machine.
* Create a snapshot. I prefer the `base` name as this is the default for the clone operation.
* For sharing export the machine to an `.ova` file.

