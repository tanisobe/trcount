# trmon
trmon is a network traffic monitoring CUI tool via SNMP.

![trmon](https://user-images.githubusercontent.com/1511945/167248808-cce3cc91-e5a5-40dd-88d1-5dd471cc34f0.jpg)

## Installation
```bash
go install github.com/tanisobe/trmon/cmd/trmon@latest
```

## Usage
```bash
trmon [OPTIONS]... AGENT...

OPTIONS
-v show version.
-c <community> snmp community in snmpv2.
-e <regexp> when regexp match I/F name or I/F descripiton, display with priority.
-i <interval> SNMP polling interval [sec].
-l <lifespan> trmon continuous operation time [sec].
```
## Example
```bash
trmon -c "my_comm" my-router my-switch
```
## Support
this tool support only snmp v2.

## License
[MIT](https://choosealicense.com/licenses/mit/)