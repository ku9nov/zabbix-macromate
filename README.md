## Overview
This is a command-line tool for managing host macros on a Zabbix server. 

It allows you to:
1. Manage host macros on a Zabbix server
2. Update existing macros
3. Authenticate with the Zabbix server

### Prerequisites

- Go version 1.15 or above
- Zabbix API credentials (username and password)
- Zabbix server link (e.g. https://zbx.example.com/zabbix/api_jsonrpc.php)
Note: You need to replace the placeholders username, password, hostname, and zbxServerLink with your actual Zabbix API credentials and Zabbix server link.

### Installing

- Clone the repository
- Run `go build` command to build the application

## Usage

To run the tool, you can use the following command:

```
go run macromate.go -username example_username -password example_password -hostname example_hostname -zbxServerLink https://zbx.example.com/zabbix/api_jsonrpc.php
```
The `-username` and `-password` flags are used to authenticate with the Zabbix server. The `-hostname` flag is used to specify the host that the macros will be associated with. And, `-zbxServerLink` flag is used to specify the link of zabbix server.

You can also run the tool with logging, by using the following command:

```
go run macromate.go -username example_username -password example_password -hostname example_hostname -zbxServerLink https://zbx.example.com/zabbix/api_jsonrpc.php -enableLogging true
```
Note: `-enableLogging` flag is added for the logging
