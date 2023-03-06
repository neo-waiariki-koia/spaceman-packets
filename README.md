# spaceman-packets

## build

`version=dev make build`

`docker build . -t spaceman-packets:dev`

## run 

### setup tcpdump container

`docker run -it --name tcpdump wbitt/network-multitool -- bash`

### run spaceman-packets

`docker run --net=container:tcpdump spaceman-packets:dev`

## capture

`tcpdump -i any -n port 8080 -vv`
