# Standing Desk

A Raspberry Pi- and Arduino-controlled standing desk: Control your [LogicData](http://www.logicdata.at)-powered standing desk remotely. The project components (hardware components, hardware wiring, software components and installation) are described in detail below.

## Hardware

### Required Components

- Jumper wires and breadboard
- Soldering equipment
- Raspberry Pi 3 (or similar)
- Arduino Micro (or similar)
- Female 7-pole DIN circular socket (to connect to handswitch)
- Male 7-pole DIN circular plug (to connect to motor controller)

### Wiring Overview

![Wiring](wiring/wiring.png)

The diagram for the wiring of DIN connections (left) was created by [Borislav Bertoldi](https://www.mikrocontroller.net/topic/373579). Colors correspond between the diagrams.

## Software

### Building Blocks

1. `DeskControl.ino` is the main program running on Arduino with the task of controlling the signal going to/from the motor controller and handswitch.
   - Copyright by [Borislav Bertoldi](https://www.mikrocontroller.net/topic/373579)
2. `desk-server.go` is a small Go program which exposes HTTP methods for controlling the desk and runs on the Raspberry Pi:
   - Adapted from [David Knezić's project](https://github.com/davidknezic/desk/blob/master/bridge.go)
   - `curl localhost:9987/up` or `/down` or `/toggle`
   - Configure `UP_HEIGHT` and `DOWN_HEIGHT` with your preferred standing and sitting height (or extend the logic yourself). Default values: `UP_HEIGHT=90` and `DOWN_HEIGHT=10`.
3. `ssh_desk_handler.py` is a small Python program which enables controlling the desk via SSH (integrate or adapt it for your own use case).
   - For simple implementation examples see [visini/stand](https://github.com/visini/stand) or [visini/timebox](https://github.com/visini/timebox)

### Installation

First, install go (tested version: go version go1.15.3 linux/arm). See instructions for Raspberry Pi [here](https://raspberrypi.stackexchange.com/questions/25956/install-golang-the-easy-way/46828#46828).

Add the following to your `~/.bashrc`:

```shell
export PATH="$PATH:$GOROOT/bin"
export GOROOT=/usr/local/go
export GOPATH=/home/pi/go/
```

Clone and build the project, and add `desk-server` to `init.d`:

```shell
git clone https://github.com/visini/desk.git ~/go/src/github.com/visini/desk
cd ~/go/src/github.com/visini/desk
go get
go build bridge.go
sudo ln -s ~/go/src/github.com/visini/desk/init.d/desk-server /etc/init.d/desk-server
sudo update-rc.d desk-server defaults
```

### Optional: Configuring `shairport-sync`

The Raspberry Pi also serves as my [shairport-sync](https://github.com/mikebrady/shairport-sync) host directly connected to my external audio interface sitting in the cable tray of my desk.

Note: If you are using an external audio interface, be sure to adjust `/usr/local/etc/shairport-sync.conf` accordingly (see also [here](https://github.com/mikebrady/shairport-sync/issues/741) for more information). To find out the hardware device `id`, run `alsamixer` and press `S`. Then configure `shairport-sync.conf` accordingly (assuming sound card device `id=1`):

```conf
output_device = "hw:1,0";
mixer_control_name = "PCM";
mixer_device = "hw:1";
```

## Attribution and Credits

I want to express my gratitude to the following people, on whose projects and effort I have heavily relied upon to implement this project:

- [A project by David Knezić](https://github.com/davidknezic/desk) which includes a [HomeBridge](https://homebridge.io) integration (i.e., allowing to control your desk from your Apple devices), from which I have adapted `desk-server.go` to expose HTTP methods.
- [A project by Borislav Bertoldi](https://www.mikrocontroller.net/topic/373579), from which I have borrowed the `DeskControl.ino` file for the Arduino, and which includes a C# application running on Windows.
- [A project by Martin Nuc](https://github.com/MartinNuc/logic-data-controller) written in Rust.
