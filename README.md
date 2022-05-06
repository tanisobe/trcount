# trmon
trmon is a network traffic monitoring tool via SNMP.

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

## Support
this tool support only snmp v2.

## License
[MIT](https://choosealicense.com/licenses/mit/)