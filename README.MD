# xd-rsync

xd-rsync is a tool to perform remote synchronisation of the XD ERP database with AWS SNS queues.
The primary purpose of this tool is to enable businesses who use XD as their ERP to create event-driven architectures, by pushing updates of certain entities to specific SNS topics.

The SNS topics could then feed into many different SQS queues in order for different systems to process the messages at their discretion.

## Setup

```bash
# Install all dependencies
go mod download

# Move to application entrypoint
cd cmd/xd-rsync

# Build & test xd-rsync
go build && ./xd-rsync

```

## Local replica

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

## But why, though?

Despite supporting different databases engines, XD's Development & CS teams do not encourage their
clients to host their databases on the cloud but rather OnPrem (on-premises).

Up to this point the only reason they can provide for being so strongly against cloud hosting, is that
it's easier for them to maintain. Guess what, **it's 2024**, the PHP era has become the triassic era for devs, and distributed systems are more common than ever.

After all, not only is it cheaper and easier to maintain distributed systems (if you're doing it right) but also scales much better than hosting everything in the same place.
