# goup: Creating Upstart Scripts

Upstart is a mechanism to handle services on Ubuntu (like SysInitV used to do). Keep in mind that this stuff requires
quite some knowledge and understanding on how processes can daemonize.

Currently there is only support for a small subset of possible options. This will be expanded as required.

Usage is as simple as creating an value of the Upstart type and call it's `CreateScript() string` method to get the
string version of the resulting upstart script. Write that to the according location.


