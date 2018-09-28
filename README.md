# AirBnB Calendar Sync Service

You can read the specification here: [Specification](spec.md)

### Usage
After running `docker-compose up -d` the following resources are available
- Postgres [http://localhost:5432](http://localhost:5432)

To build the binary run `make`. The API will be accessible via [http://localhost:9090](http://localhost:9090)

### API Endpoints

- `/calendar/ical.ics` should be consumed by AirBnB

An endpoint could also be provided to store existing events in the AirBnB export calendar.
