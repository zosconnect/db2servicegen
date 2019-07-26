# db2servicegen

Sample program to create z/OS Connect EE service projects for Db2 RESTful Services.

## Usage

```
db2servicegen --server http://example.mainframe.com:42100 --user admin --password adminpass
```

* `--server` : The hostname and port of the Db2 server
* `--user` : The UserId of the user that can retrieve all the Db2 services
* `--password` : The password for the UserId specified in the `--user` parameter

## How it works

The program does the following to create the files required to create SARs for the Db2 services.

1. Connect to the Db2 server and retrieve the list of services from the `/services` endpoint.
1. Iterate through the services returned and get the information about each one.
1. Extract the request and response schemas for the Db2 service and remove any references to `null` types which are unsupported by z/OS Connect EE.
1. Create a `service.properties` file with the details of the service.
1. Save the files in a new directory, named the same as the Db2 service, in the current working directory.

SAR files can then be built using the z/OS Connect EE build toolkit by going into the directory and running 

```
zconbt -p service.properties -f serviceName.sar
```

## License

```
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```