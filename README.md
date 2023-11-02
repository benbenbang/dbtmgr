# `dbtmgr` CLI Tool

`dbtmgr` is a command-line interface (CLI) tool designed to manage and synchronize database table schema states for development teams using `dbt` (data build tool). It ensures that schema changes are tracked and versioned, and enables developers to acquire exclusive locks on the state during updates to prevent concurrent conflicts.



## Features

- **State Management:** Track and synchronize table schema states across your development team.
- **Lock Mechanism:** Safeguard state changes with a lock mechanism using an S3 bucket to avoid concurrent update issues.
- **State Synchronization:** Provides commands to refresh local states with remote states and push local changes to the remote state.
- **Easy Configuration:** Simple setup process with configuration options for AWS credentials and S3 bucket details.



## Installation

To install `dbtmgr`, you can use `go get`:

```bash
go get github.com/yourusername/dbtmgr
```

Alternatively, you can clone the repository and build from source:

```bash
git clone https://github.com/yourusername/dbtmgr.git
cd dbtmgr
go build .
```



## Usage

Before using `dbtmgr`, ensure you have configured your AWS credentials and have the necessary permissions to read from and write to the specified S3 bucket.

### Common Commands

- `dbtmgr acquire`: Acquires a lock on the state file within the S3 bucket to prevent others from making concurrent state changes.
- `dbtmgr release`: Releases the lock on the state file within the S3 bucket.
- `dbtmgr refresh`: Pulls the latest state from the S3 bucket to your local environment.
- `dbtmgr sync`: Pushes the local state changes to the S3 bucket.

### Examples

Acquire / Release / Refresh / Sync a lock:

```bash
dbtmgr lock acquire
dbtmgr lock release
dbtmgr lock refresh
dbtmgr lock sync
```



## Contributing

Contributions to `dbtmgr` are welcome! Please fork the repository and submit a pull request with your changes or improvements.

## License

`dbtmgr` is released under the Apache License. See the LICENSE file for more details.
