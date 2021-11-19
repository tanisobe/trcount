# trcount
trcount is a network traffic monitoring tool via SNMP.

## Installation
```bash
go install github.com/tanisobe/trcount/cmd/trcount
```

## Usage
```bash
trcount [OPTIONS]... AGENT...

OPTIONS
-c <community> snmp community in snmpv2.
-e <regexp> when regexp match I/F name or I/F descripiton, display with priority.
```

## Support
this tool support only snmp v2.

## License
[MIT](https://choosealicense.com/licenses/mit/)