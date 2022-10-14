# service-intesis

a partial golang port of the HomeAssistant Intesis Home controller located here
https://github.com/jnimmo/pyIntesisHome

this version allows for:

* querying device state
* setting various values to HVAC entities
* monitoring the state & exporting the metrics out to prometheus

# notice

at no point am I associated with the Intesis product at all, nor do I have the
spec of the API, I'm just following what's already out there

for license information have a look at `LICENSE`

# testing

`go test ./...`

# running

list your devices

`go run cmd/service-intesis.go -username x -password y`

once you get the device list query it's status

`go run cmd/service-intesis.go -username x -password y -device z`

you can then command the device to change state

`go run cmd/service-intesis.go -username x -password y -device z -set thing -value value`

where:

* thing can be either the uid of a control or it's name
* value can be a named enum or it's actual value

finally you can continuously monitor the state & export the metrics out for prometheus scraping

`go run cmd/service-intesis.go -monitor -device x`

in `-monitor` we expect secrets to be located at `/.secrets/creds.yaml` containing the username & password

# building

`./package.sh`

# TODO

loads