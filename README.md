dp-dataset-api
==================
A ONS API used to navigate datasets which are published.

#### Postgres
* Run ```brew install mongo```
* Run ```sudo mkdir -p /data/db```
* Run ```sudo chmod 777 /data/db```
* Run ```mongod &```
* Run ```./scripts/InitDatabase.sh```

### Configuration

<<<<<<< HEAD
| Environment variable             | Default                              | Description
| -------------------------------- | -------------------------------------| -----------
| BIND_ADDR                        | :22000                               | The host and port to bind to
| MONGODB_BIND_ADDR                | localhost:27017                      | The MongoDB bind address
| MONGODB_DATABASE                 | datasets                             | The MongoDB dataset database
| MONGODB_COLLECTION               | datasets                             | MongoDB collection
| SECRET_KEY                       | FD0108EA-825D-411C-9B1D-41EF7727F465 | A secret key used authentication
| CODE_LIST_API_URL                | http://localhost:22400               | The host name for the CodeList API
| DATASET_API_URL                  | http://localhost:22000               | The host name for the Dataset API
| GRACEFUL_SHUTDOWN_TIMEOUT        | 5s                                   | The graceful shutdown timeout in seconds
| WEBSITE_URL                | http://localhost:20000               | The host name for the website
| KAFKA_ADDR                       | "localhost:9092"                     | The list of kafka hosts
| GENERATE_DOWNLOADS_TOPIC         | "filter-job-submitted"               | The topic to send generate full dataset version downloads to

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details
