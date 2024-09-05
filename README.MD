# xd-rsync

xd-rsync is a tool to perform remote synchronisation of the XD ERP database with AWS SNS queues.
The primary purpose of this tool is to enable businesses who use XD as their ERP to create event-driven architectures, by pushing updates of certain entities to specific SNS topics.

The SNS topics could then feed into many different SQS queues in order for different systems to process the messages at their discretion.

## Usage

### Prerequisites

1. Have an instance of XD ERP running alongside its local database.
2. Create `config.json` in the same folder where xd-rsync is going to be executed.

### Config example

Customise your configuration file (`config.json`) according to your needs.

[Datadog ingest URLs can be found on their documentation.](https://docs.datadoghq.com/logs/log_collection/?tab=host#supported-endpoints)

```json
{
  "environment": "development",
  "awsRegion": "eu-west-2",
  // Database DSN for the xd-rsync client to be able to connect
  "dsn": "root:root@tcp(localhost:3306)/xd?charset=utf8mb4&parseTime=True&loc=Local",
  "queues": {
    // SNS topic for product updates to be published
    "productUpdatesSnsQueueArn": ""
  },
  "datadog": {
    // Datadog custom host
    "ingestHost": "http-intake.logs.datadoghq.eu",
    // Datadog API key
    "apiKey": "<INSERT_DATADOG_KEY_HERE>",

    // Insert fields that all events should contain
    "eventBaseFields": {
      "customEventProperty": "value"
    }
  },
  "syncFrequency": "5m" // Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
}
```

## Development

### Setup

```bash
# Install all dependencies
go mod download

# Move to application entrypoint
cd cmd/xd-rsync

# Build & test xd-rsync
go build && ./xd-rsync

```

### Local replica

The first step is to perform a database dump from XD database. After moving it into the folder [dumps/](/dumps/), run:

```bash
# Start Docker-contained MySQL service
docker compose up

# Stop database container
docker compose down

# Reset database container by removing named volumes
docker compose down -v && docker compose up --force-recreate

# Run dump SQL script against container
mysql --max_allowed_packet=256M -h localhost -u root --protocol=tcp --password=root xd < ./dumps/dumpname.sql
```

## Why use xd-rsync?

Despite supporting different databases engines, XD's Development & CS teams do not encourage their
clients to host their databases on the cloud but rather OnPrem (on-premises).

Up to this point the only reason they can provide for being so strongly against cloud hosting, is that
it's easier for them to maintain. Guess what, **it's 2024**, the PHP era has become the triassic era for devs, and distributed systems are more common than ever.

After all, not only is it cheaper and easier to maintain distributed systems (if you're doing it right) but also scales much better than hosting everything in the same place.

### How does it work

The product changes are tracked using a combination of fields:

| Change                                                    | DB Field                                                    |
| --------------------------------------------------------- | ----------------------------------------------------------- |
| Product details change                                    | Product table's sync timestamp: `items.SyncStamp`           |
| Stock entry details change (not updated on entrance/exit) | Item stock table's sync timestamp: `itemstock.SyncStamp`    |
| Stock entrance movement                                   | Item stock table's sync timestamp: `itemstock.LastEntrance` |
| Stock exit movement                                       | Item stock table's sync timestamp: `itemstock.LastExit`     |

Filtering the products whose tracking fields have changed since the last update, xd-rsync is able to then push the updates to the SNS topic provided.
