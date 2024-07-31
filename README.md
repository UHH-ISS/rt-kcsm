## Usage

### Command Line Options

**Ingesting data**

You have two options for getting the data in. Either via file or TCP server:
- `--file /my/file` File path of your log file
- `--listen 0.0.0.0:9000` TCP Server address that can be connected to using Tenzir's [TCP client connector](https://docs.tenzir.com/connectors/tcp)

In addition, you also need to supply the `--transport TYPE` option with either:

- `file` when using the `--file` option (**default**)
- `tcp` when using the `--listen` option


Using the TCP server has the advantage of enabling data streaming, thereby generating graphs in real-time.

**Data ingest type**

You can choose between different data types of your ingested data using the `--reader TYPE` option:
- `zeek`
- `suricata`
- `ocsf`


**Web UI**

By default the RT-KCSM does not expose the web interface. To view it, you must specify the address of the HTTP server:
- `--server 0.0.0.0:8080`

You can visit the web UI at [http://localhost:8080/web/](http://localhost:8080/web/)

For configuring high-risks target/hosts go to the `Configure Hosts` section.

**Example using Docker**

Default options for Docker container.
```
docker run git.informatik.uni-hamburg.de:4567/iss/projects/sovereign/rt-kcsm/open-source:latest --reader suricata --transport tcp --listen :9000 --server :8080
```