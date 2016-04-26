# ghprs & ghprs2iflx
The `ghprs` command line utility fetches PR info for a respository and saves it to disk in JSON encoded format. This makes it easy for other utilities to read and analyze.

The `ghprs2iflx` command line utility reads a JSON file written by `ghprs` and stores the PR info in InfluxDB. A text file containing a list of GitHub logins for the project's team members can optionally be specified. If a list of team members is specified, PRs **not** from one of these team members will be tagged as `community = 'yes'`. Queries can then be run to determine if community contributions are trending up or down.

# usage
Run either command with the `-help` command line argument to get a list of options.
