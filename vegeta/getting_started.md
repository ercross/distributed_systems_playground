# Getting Started with Vegeta 

This project intentionally starts with a **fragile distributed system**.

Your goal today is **not performance tuning**.
Your goal is to **observe failure under load** and learn how distributed systems break *before* they look busy.

Vegeta is the tool we’ll use to apply pressure.


## What Vegeta Is (and Isn’t)

Vegeta is an HTTP load generator.

It:
- sends requests at a controlled rate
- records latency, throughput, and errors
- shows how systems behave under stress

Vegeta is not:
- a benchmarking tool
- a performance scorecard
- a replacement for observability

Here, it is a failure amplifier.


## Installation

### macOS (Homebrew)
```bash
brew install vegeta
````

### Linux (Download binary)

```bash
curl -LO https://github.com/tsenart/vegeta/releases/latest/download/vegeta_linux_amd64.tar.gz
tar -xzf vegeta_linux_amd64.tar.gz
sudo mv vegeta /usr/local/bin
```

Verify:

```bash
vegeta version
```

### v0.1.0 Repository Layout (Vegeta)

```
vegeta/
├── targets/
│   ├── dsp.txt        # Request definitions
│   └── dsp_body.json  # POST body
└── results/
    ├── 50rps.bin
    ├── 100rps.bin
    ├── 200rps.bin
    └── 1000rps.bin
```

* `targets/` defines **what** we send
* `results/` stores **what happened**


### Targets Configuration

#### `targets/dsp.txt`

Example:

```
POST http://localhost:15002/orders
Content-Type: application/json
@dsp_body.json
```

This tells Vegeta:

* HTTP method
* endpoint
* headers
* request body file

Vegeta will reuse this payload for all requests.



