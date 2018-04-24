# reconcile
Reconcile takes a CSV file as input and will check Tetration to see if the hosts have checked in. This is useful in deployment scenarios to match actual vs expected sensors.

## Building
`go build .`

## Usage
`./reconcile -input ./testdata/records.csv`

input is the csv file, the csv file format is:
```
hostname,ip
host1,1.1.1.1
host1,1.1.1.2
host2,2.2.2.1
```
Note: first line will be treated as a description and will not be parsed.


## License

This project is licensed under the GPLv3 License - see the [LICENSE](LICENSE) file for details
