# service-intesis

a partial golang port of the HomeAssistant Intesis Home controller located here
https://github.com/jnimmo/pyIntesisHome

it's also a k8s deployment which exports intesis home based HVAC state to prometheus
it's also meant to support controlling the HVAC (but is a work in progress)

I'm using it to monitor & control my HVAC via k8s on some Raspberry Pi 3's

# testing

`go test ./...`

# running

`go run . -username x -password x -device x`

where:
    * username is the intesis cloud username
    * password is ...
    * device is the HVAC device which has the intesis plugged into it

without `-device` will list the devices configured

# building

`./package.sh`

# TODO

* `-monitor` flag
* prometheus metrics exporting
* implementation of command setting
* moving the cli to a package of it's own